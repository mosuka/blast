package storage

import (
	"os"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/mapping"
	bleveindex "github.com/blevesearch/bleve_index_api"
	_ "github.com/mosuka/blast/builtin"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
)

type Index struct {
	indexMapping *mapping.IndexMappingImpl
	logger       *zap.Logger

	index bleve.Index
}

func NewIndex(dir string, indexMapping *mapping.IndexMappingImpl, logger *zap.Logger) (*Index, error) {
	var index bleve.Index

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// create new index
		index, err = bleve.NewUsing(dir, indexMapping, scorch.Name, scorch.Name, nil)
		if err != nil {
			logger.Error("failed to create index", zap.String("dir", dir), zap.Error(err))
			return nil, err
		}
	} else {
		// open existing index
		index, err = bleve.OpenUsing(dir, map[string]interface{}{
			"create_if_missing": false,
			"error_if_exists":   false,
		})
		if err != nil {
			logger.Error("failed to open index", zap.String("dir", dir), zap.Error(err))
			return nil, err
		}
	}

	return &Index{
		index:        index,
		indexMapping: indexMapping,
		logger:       logger,
	}, nil
}

func (i *Index) Close() error {
	if err := i.index.Close(); err != nil {
		i.logger.Error("failed to close index", zap.Error(err))
		return err
	}

	return nil
}

func (i *Index) Get(id string) (map[string]interface{}, error) {
	doc, err := i.index.Document(id)
	if err != nil {
		i.logger.Error("failed to get document", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	if doc == nil {
		err := errors.ErrNotFound
		i.logger.Debug("document does not found", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	fields := make(map[string]interface{}, 0)
	doc.VisitFields(func(field bleveindex.Field) {
		var v interface{}
		switch field := field.(type) {
		case bleveindex.TextField:
			v = field.Text()
		case bleveindex.NumericField:
			n, err := field.Number()
			if err == nil {
				v = n
			}
		case bleveindex.DateTimeField:
			d, err := field.DateTime()
			if err == nil {
				v = d.Format(time.RFC3339Nano)
			}
		}
		existing, existed := fields[field.Name()]
		if existed {
			switch existing := existing.(type) {
			case []interface{}:
				fields[field.Name()] = append(existing, v)
			case interface{}:
				arr := make([]interface{}, 2)
				arr[0] = existing
				arr[1] = v
				fields[field.Name()] = arr
			}
		} else {
			fields[field.Name()] = v
		}
	})

	return fields, nil
}

func (i *Index) Search(searchRequest *bleve.SearchRequest) (*bleve.SearchResult, error) {
	searchResult, err := i.index.Search(searchRequest)
	if err != nil {
		i.logger.Error("failed to search documents", zap.Any("search_request", searchRequest), zap.Error(err))
		return nil, err
	}

	return searchResult, nil
}

func (i *Index) Index(id string, fields map[string]interface{}) error {
	if err := i.index.Index(id, fields); err != nil {
		i.logger.Error("failed to index document", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

func (i *Index) Delete(id string) error {
	if err := i.index.Delete(id); err != nil {
		i.logger.Error("failed to delete document", zap.String("id", id), zap.Error(err))
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
			err := errors.ErrNil
			i.logger.Error("missing id", zap.Error(err))
			continue
		}
		fields, ok := doc["fields"].(map[string]interface{})
		if !ok {
			err := errors.ErrNil
			i.logger.Error("missing fields", zap.Error(err))
			continue
		}

		if err := batch.Index(id, fields); err != nil {
			i.logger.Error("failed to index document in batch", zap.String("id", id), zap.Error(err))
			continue
		}
		count++
	}

	err := i.index.Batch(batch)
	if err != nil {
		i.logger.Error("failed to index documents", zap.Int("count", count), zap.Error(err))
		return count, err
	}

	if count <= 0 {
		err := errors.ErrNoUpdate
		i.logger.Error("no documents updated", zap.Any("count", count), zap.Error(err))
		return count, err
	}

	return count, nil
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
		i.logger.Error("failed to delete documents", zap.Int("count", count), zap.Error(err))
		return count, err
	}

	return count, nil
}

func (i *Index) Mapping() *mapping.IndexMappingImpl {
	return i.indexMapping
}

func (i *Index) Stats() map[string]interface{} {
	return i.index.StatsMap()
}

func (i *Index) SnapshotItems() <-chan *protobuf.Document {
	ch := make(chan *protobuf.Document, 1024)

	go func() {
		idx, err := i.index.Advanced()
		if err != nil {
			i.logger.Error("failed to get index", zap.Error(err))
			return
		}

		ir, err := idx.Reader()
		if err != nil {
			i.logger.Error("failed to get index reader", zap.Error(err))
			return
		}

		docCount := 0

		dr, err := ir.DocIDReaderAll()
		if err != nil {
			i.logger.Error("failed to get doc ID reader", zap.Error(err))
			return
		}
		for {
			//if dr == nil {
			//	i.logger.Error(err.Error())
			//	break
			//}
			id, err := dr.Next()
			if id == nil {
				i.logger.Debug("finished to read all document IDs")
				break
			} else if err != nil {
				i.logger.Warn("failed to get doc ID", zap.Error(err))
				continue
			}

			// get original document
			fieldsBytes, err := i.index.GetInternal(id)
			if err != nil {
				i.logger.Warn("failed to get doc fields bytes", zap.String("id", string(id)), zap.Error(err))
				continue
			}

			doc := &protobuf.Document{
				Id:     string(id),
				Fields: fieldsBytes,
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
