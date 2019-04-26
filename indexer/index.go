// Copyright (c) 2019 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexer

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/ptypes/any"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
)

type Index struct {
	index bleve.Index

	indexMapping     *mapping.IndexMappingImpl
	indexType        string
	indexStorageType string

	logger *log.Logger
}

func NewIndex(dir string, indexMapping *mapping.IndexMappingImpl, indexType string, indexStorageType string, logger *log.Logger) (*Index, error) {
	bleve.SetLog(logger)

	var index bleve.Index
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// create new index
		index, err = bleve.NewUsing(dir, indexMapping, indexType, indexStorageType, nil)
		if err != nil {
			return nil, err
		}
	} else {
		// open existing index
		index, err = bleve.OpenUsing(dir, map[string]interface{}{
			"create_if_missing": false,
			"error_if_exists":   false,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Index{
		index:            index,
		indexMapping:     indexMapping,
		indexType:        indexType,
		indexStorageType: indexStorageType,
		logger:           logger,
	}, nil
}

func (i *Index) Close() error {
	err := i.index.Close()
	if err != nil {
		return err
	}

	return nil
}

func (i *Index) Get(id string) (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		i.logger.Printf("[DEBUG] get %s %f", id, float64(time.Since(start))/float64(time.Second))
	}()

	fieldsBytes, err := i.index.GetInternal([]byte(id))
	if err != nil {
		return nil, err
	}
	if len(fieldsBytes) <= 0 {
		return nil, blasterrors.ErrNotFound
	}

	// bytes -> map[string]interface{}
	var fieldsMap map[string]interface{}
	err = json.Unmarshal(fieldsBytes, &fieldsMap)
	if err != nil {
		return nil, err
	}

	return fieldsMap, nil
}

func (i *Index) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	start := time.Now()
	defer func() {
		rb, _ := json.Marshal(request)
		i.logger.Printf("[DEBUG] search %s %f", rb, float64(time.Since(start))/float64(time.Second))
	}()

	result, err := i.index.Search(request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *Index) Index(id string, fields map[string]interface{}) error {
	start := time.Now()
	defer func() {
		i.logger.Printf("[DEBUG] index %s %v %f", id, fields, float64(time.Since(start))/float64(time.Second))
	}()

	// index
	i.logger.Printf("[DEBUG] index %s, %v", id, fields)
	err := i.index.Index(id, fields)
	if err != nil {
		return err
	}
	i.logger.Printf("[DEBUG] indexed %s, %v", id, fields)

	// map[string]interface{} -> bytes
	fieldsBytes, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	// set original document
	err = i.index.SetInternal([]byte(id), fieldsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (i *Index) Delete(id string) error {
	start := time.Now()
	defer func() {
		i.logger.Printf("[DEBUG] delete %s %f", id, float64(time.Since(start))/float64(time.Second))
	}()

	err := i.index.Delete(id)
	if err != nil {
		return err
	}

	// delete original document
	err = i.index.SetInternal([]byte(id), nil)
	if err != nil {
		return err
	}

	return nil
}

func (i *Index) Config() (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		i.logger.Printf("[DEBUG] stats %f", float64(time.Since(start))/float64(time.Second))
	}()

	indexConfig := map[string]interface{}{
		"index_mapping":      i.indexMapping,
		"index_type":         i.indexType,
		"index_storage_type": i.indexStorageType,
	}

	return indexConfig, nil
}

func (i *Index) Stats() (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		i.logger.Printf("[DEBUG] stats %f", float64(time.Since(start))/float64(time.Second))
	}()

	stats := i.index.StatsMap()

	return stats, nil
}

func (i *Index) SnapshotItems() <-chan *pbindex.Document {
	ch := make(chan *pbindex.Document, 1024)

	go func() {
		idx, _, err := i.index.Advanced()
		if err != nil {
			i.logger.Printf("[ERR] %v", err)
			return
		}

		r, err := idx.Reader()
		if err != nil {
			i.logger.Printf("[ERR] %v", err)
			return
		}

		docCount := 0

		dr, err := r.DocIDReaderAll()
		for {
			id, err := dr.Next()
			if id == nil {
				i.logger.Print("[DEBUG] finished to read all document ids")
				break
			} else if err != nil {
				i.logger.Printf("[WARN] %v", err)
				continue
			}

			// get original document
			fieldsBytes, err := i.index.GetInternal(id)

			// bytes -> map[string]interface{}
			var fieldsMap map[string]interface{}
			err = json.Unmarshal([]byte(fieldsBytes), &fieldsMap)
			if err != nil {
				i.logger.Printf("[ERR] %v", err)
				break
			}
			i.logger.Printf("[DEBUG] %v", fieldsMap)

			// map[string]interface{} -> Any
			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
			if err != nil {
				i.logger.Printf("[ERR] %v", err)
				break
			}

			doc := &pbindex.Document{
				Id:     string(id),
				Fields: fieldsAny,
			}

			ch <- doc

			docCount = docCount + 1
		}

		i.logger.Print("[DEBUG] finished to write all documents to channel")
		ch <- nil

		i.logger.Printf("[INFO] snapshot total %d documents", docCount)

		return
	}()

	return ch
}
