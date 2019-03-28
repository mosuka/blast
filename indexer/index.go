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
	index  bleve.Index
	logger *log.Logger
}

func NewIndex(dir string, indexMapping *mapping.IndexMappingImpl, indexStorageType string, logger *log.Logger) (*Index, error) {
	bleve.SetLog(logger)

	var index bleve.Index
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// create new index
		index, err = bleve.NewUsing(dir, indexMapping, bleve.Config.DefaultIndexType, indexStorageType, nil)
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
		index:  index,
		logger: logger,
	}, nil
}

func (b *Index) Close() error {
	err := b.index.Close()
	if err != nil {
		return err
	}

	return nil
}

func (b *Index) Get(id string) (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] get %s %f", id, float64(time.Since(start))/float64(time.Second))
	}()

	fieldsBytes, err := b.index.GetInternal([]byte(id))
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

func (b *Index) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	start := time.Now()
	defer func() {
		rb, _ := json.Marshal(request)
		b.logger.Printf("[DEBUG] search %s %f", rb, float64(time.Since(start))/float64(time.Second))
	}()

	result, err := b.index.Search(request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *Index) Index(id string, fields map[string]interface{}) error {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] index %s %v %f", id, fields, float64(time.Since(start))/float64(time.Second))
	}()

	// index
	b.logger.Printf("[DEBUG] index %s, %v", id, fields)
	err := b.index.Index(id, fields)
	if err != nil {
		return err
	}
	b.logger.Printf("[DEBUG] indexed %s, %v", id, fields)

	// map[string]interface{} -> bytes
	fieldsBytes, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	// set original document
	err = b.index.SetInternal([]byte(id), fieldsBytes)
	if err != nil {
		return err
	}

	return nil
}

func (b *Index) Delete(id string) error {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] delete %s %f", id, float64(time.Since(start))/float64(time.Second))
	}()

	err := b.index.Delete(id)
	if err != nil {
		return err
	}

	// delete original document
	err = b.index.SetInternal([]byte(id), nil)
	if err != nil {
		return err
	}

	return nil
}

func (b *Index) Stats() (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] stats %f", float64(time.Since(start))/float64(time.Second))
	}()

	stats := b.index.StatsMap()

	return stats, nil
}

func (b *Index) SnapshotItems() <-chan *pbindex.Document {
	ch := make(chan *pbindex.Document, 1024)

	go func() {
		i, _, err := b.index.Advanced()
		if err != nil {
			b.logger.Printf("[ERR] %v", err)
			return
		}

		r, err := i.Reader()
		if err != nil {
			b.logger.Printf("[ERR] %v", err)
			return
		}

		docCount := 0

		dr, err := r.DocIDReaderAll()
		for {
			id, err := dr.Next()
			if id == nil {
				b.logger.Print("[DEBUG] finished to read all document ids")
				break
			} else if err != nil {
				b.logger.Printf("[WARN] %v", err)
				continue
			}

			// get original document
			fieldsBytes, err := b.index.GetInternal(id)

			// bytes -> map[string]interface{}
			var fieldsMap map[string]interface{}
			err = json.Unmarshal([]byte(fieldsBytes), &fieldsMap)
			if err != nil {
				b.logger.Printf("[ERR] %v", err)
				break
			}
			b.logger.Printf("[DEBUG] %v", fieldsMap)

			// map[string]interface{} -> Any
			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
			if err != nil {
				b.logger.Printf("[ERR] %v", err)
				break
			}

			doc := &pbindex.Document{
				Id:     string(id),
				Fields: fieldsAny,
			}

			ch <- doc

			docCount = docCount + 1
		}

		b.logger.Print("[DEBUG] finished to write all documents to channel")
		ch <- nil

		b.logger.Printf("[INFO] snapshot total %d documents", docCount)

		return
	}()

	return ch
}
