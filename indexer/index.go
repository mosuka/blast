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
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"go.uber.org/zap"
)

type Index struct {
	indexMapping     *mapping.IndexMappingImpl
	indexType        string
	indexStorageType string
	logger           *zap.Logger

	index bleve.Index
}

func NewIndex(dir string, indexMapping *mapping.IndexMappingImpl, indexType string, indexStorageType string, logger *zap.Logger) (*Index, error) {
	//bleve.SetLog(logger)

	var index bleve.Index
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
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
		i.logger.Error(err.Error())
		return err
	}

	return nil
}

func (i *Index) Get(id string) (map[string]interface{}, error) {
	doc, err := i.index.Document(id)
	if err != nil {
		i.logger.Error(err.Error())
		return nil, err
	}
	if doc == nil {
		return nil, errors.ErrNotFound
	}

	fields := make(map[string]interface{}, 0)
	for _, f := range doc.Fields {
		var v interface{}
		switch field := f.(type) {
		case *document.TextField:
			v = string(field.Value())
		case *document.NumericField:
			n, err := field.Number()
			if err == nil {
				v = n
			}
		case *document.DateTimeField:
			d, err := field.DateTime()
			if err == nil {
				v = d.Format(time.RFC3339Nano)
			}
		}
		existing, existed := fields[f.Name()]
		if existed {
			switch existing := existing.(type) {
			case []interface{}:
				fields[f.Name()] = append(existing, v)
			case interface{}:
				arr := make([]interface{}, 2)
				arr[0] = existing
				arr[1] = v
				fields[f.Name()] = arr
			}
		} else {
			fields[f.Name()] = v
		}
	}

	return fields, nil
}

func (i *Index) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := i.index.Search(request)
	if err != nil {
		i.logger.Error(err.Error())
		return nil, err
	}

	return result, nil
}

func (i *Index) Index(doc *index.Document) error {
	_, err := i.BulkIndex([]*index.Document{doc})
	if err != nil {
		i.logger.Error(err.Error())
		return err
	}

	return nil
}

func (i *Index) BulkIndex(docs []*index.Document) (int, error) {
	batch := i.index.NewBatch()

	count := 0

	for _, doc := range docs {
		fieldsIntr, err := protobuf.MarshalAny(doc.Fields)
		if err != nil {
			i.logger.Error(err.Error(), zap.Any("doc", doc))
			continue
		}
		err = batch.Index(doc.Id, *fieldsIntr.(*map[string]interface{}))
		if err != nil {
			i.logger.Error(err.Error())
			continue
		}
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
	return map[string]interface{}{
		"index_mapping":      i.indexMapping,
		"index_type":         i.indexType,
		"index_storage_type": i.indexStorageType,
	}, nil
}

func (i *Index) Stats() (map[string]interface{}, error) {
	return i.index.StatsMap(), nil
}

func (i *Index) SnapshotItems() <-chan *index.Document {
	ch := make(chan *index.Document, 1024)

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

			doc := &index.Document{
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
