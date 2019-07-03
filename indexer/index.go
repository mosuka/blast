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
	"os"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
)

type Index struct {
	index bleve.Index

	indexConfig map[string]interface{}

	logger *zap.Logger
}

func NewIndex(dir string, indexConfig map[string]interface{}, logger *zap.Logger) (*Index, error) {
	//bleve.SetLog(logger)

	var index bleve.Index
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// default index mapping
		indexMapping := bleve.NewIndexMapping()

		// index mapping from config
		indexMappingIntr, ok := indexConfig["index_mapping"]
		if ok {
			if indexMappingIntr != nil {
				indexMappingBytes, err := json.Marshal(indexMappingIntr)
				if err != nil {
					logger.Error(err.Error())
					return nil, err
				}
				err = json.Unmarshal(indexMappingBytes, indexMapping)
				if err != nil {
					logger.Error(err.Error())
					return nil, err
				}
			}
		} else {
			logger.Error("missing index mapping")
		}

		indexType, ok := indexConfig["index_type"].(string)
		if !ok {
			logger.Error("missing index type")
			indexType = bleve.Config.DefaultIndexType
		}

		indexStorageType, ok := indexConfig["index_storage_type"].(string)
		if !ok {
			logger.Error("missing index storage type")
			indexStorageType = bleve.Config.DefaultKVStore
		}

		// create new index
		index, err = bleve.NewUsing(dir, indexMapping, indexType, indexStorageType, nil)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	} else {
		// open existing index
		index, err = bleve.OpenUsing(dir, map[string]interface{}{
			"create_if_missing": false,
			"error_if_exists":   false,
		})
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}

	return &Index{
		index:       index,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (i *Index) Close() error {
	err := i.index.Close()
	if err != nil {
		i.logger.Error(err.Error())
		return err
	}

	return nil
}

func (i *Index) Get(id string) (map[string]interface{}, error) {
	fieldsBytes, err := i.index.GetInternal([]byte(id))
	if err != nil {
		i.logger.Error(err.Error())
		return nil, err
	}
	if len(fieldsBytes) <= 0 {
		i.logger.Error(blasterrors.ErrNotFound.Error())
		return nil, blasterrors.ErrNotFound
	}

	var fieldsMap map[string]interface{}
	err = json.Unmarshal(fieldsBytes, &fieldsMap)
	if err != nil {
		i.logger.Error(err.Error())
		return nil, err
	}

	return fieldsMap, nil
}

func (i *Index) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := i.index.Search(request)
	if err != nil {
		i.logger.Error(err.Error())
		return nil, err
	}

	return result, nil
}

func (i *Index) Index(id string, fields map[string]interface{}) error {
	doc := map[string]interface{}{
		"id":     id,
		"fields": fields,
	}
	_, err := i.BulkIndex([]map[string]interface{}{doc})
	if err != nil {
		i.logger.Error(err.Error())
		return err
	}

	return nil
}

func (i *Index) BulkIndex(docs []map[string]interface{}) (int, error) {
	batch := i.index.NewBatch()

	count := 0

	for _, doc := range docs {
		id, ok := doc["id"].(string)
		if !ok {
			i.logger.Error("missing document id")
			continue
		}
		fields, ok := doc["fields"].(map[string]interface{})
		if !ok {
			i.logger.Error("missing document fields")
			continue
		}
		err := batch.Index(id, fields)
		if err != nil {
			i.logger.Error(err.Error())
			continue
		}

		// set original document
		fieldsBytes, err := json.Marshal(fields)
		if err != nil {
			i.logger.Error(err.Error())
			continue
		}
		batch.SetInternal([]byte(id), fieldsBytes)

		count++
	}

	err := i.index.Batch(batch)
	if err != nil {
		i.logger.Error(err.Error())
		return -1, err
	}

	return count, nil
}

func (i *Index) Delete(id string) error {
	_, err := i.BulkDelete([]string{id})
	if err != nil {
		i.logger.Error(err.Error())
		return err
	}

	return nil
}

func (i *Index) BulkDelete(ids []string) (int, error) {
	batch := i.index.NewBatch()

	count := 0

	for _, id := range ids {
		batch.Delete(id)

		// delete original document
		batch.SetInternal([]byte(id), nil)

		count++
	}

	err := i.index.Batch(batch)
	if err != nil {
		i.logger.Error(err.Error())
		return -1, err
	}

	return count, nil
}

func (i *Index) Config() (map[string]interface{}, error) {
	return i.indexConfig, nil
}

func (i *Index) Stats() (map[string]interface{}, error) {
	return i.index.StatsMap(), nil
}

func (i *Index) SnapshotItems() <-chan *protobuf.Document {
	ch := make(chan *protobuf.Document, 1024)

	go func() {
		idx, _, err := i.index.Advanced()
		if err != nil {
			i.logger.Error(err.Error())
			return
		}

		r, err := idx.Reader()
		if err != nil {
			i.logger.Error(err.Error())
			return
		}

		docCount := 0

		dr, err := r.DocIDReaderAll()
		for {
			if dr == nil {
				i.logger.Error(err.Error())
				break
			}
			id, err := dr.Next()
			if id == nil {
				i.logger.Debug("finished to read all document ids")
				break
			} else if err != nil {
				i.logger.Warn(err.Error())
				continue
			}

			// get original document
			fieldsBytes, err := i.index.GetInternal(id)

			// bytes -> map[string]interface{}
			var fieldsMap map[string]interface{}
			err = json.Unmarshal([]byte(fieldsBytes), &fieldsMap)
			if err != nil {
				i.logger.Error(err.Error())
				break
			}

			// map[string]interface{} -> Any
			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
			if err != nil {
				i.logger.Error(err.Error())
				break
			}

			doc := &protobuf.Document{
				Id:     string(id),
				Fields: fieldsAny,
			}

			ch <- doc

			docCount = docCount + 1
		}

		i.logger.Debug("finished to write all documents to channel")
		ch <- nil

		i.logger.Info("finished to snapshot", zap.Int("count", docCount))

		return
	}()

	return ch
}
