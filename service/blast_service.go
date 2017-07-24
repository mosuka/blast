//  Copyright (c) 2017 Minoru Osuka
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

package service

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	_ "github.com/mosuka/blast/dependency"
	"github.com/mosuka/blast/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os"
	"time"
)

type BlastService struct {
	Path         string
	IndexMapping *mapping.IndexMappingImpl
	IndexType    string
	Kvstore      string
	Kvconfig     map[string]interface{}
	Index        bleve.Index
}

func NewBlastService(path string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}) *BlastService {
	return &BlastService{
		Path:         path,
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
		Index:        nil,
	}
}

func (s *BlastService) OpenIndex() error {
	_, err := os.Stat(s.Path)
	if os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"path": s.Path,
		}).Info("index does not exist")

		s.Index, err = bleve.NewUsing(s.Path, s.IndexMapping, s.IndexType, s.Kvstore, s.Kvconfig)
		if err == nil {
			log.WithFields(log.Fields{
				"path":         s.Path,
				"indexMapping": s.IndexMapping,
				"indexType":    s.IndexType,
				"kvstore":      s.Kvstore,
				"kvconfig":     s.Kvconfig,
			}).Info("succeeded in creating index")
		} else {
			log.WithFields(log.Fields{
				"path":         s.Path,
				"indexMapping": s.IndexMapping,
				"indexType":    s.IndexType,
				"kvstore":      s.Kvstore,
				"kvconfig":     s.Kvconfig,
				"err":          err,
			}).Error("failed to create index")
		}
	} else {
		log.WithFields(log.Fields{
			"path": s.Path,
		}).Info("index exists")

		s.Index, err = bleve.OpenUsing(s.Path, s.Kvconfig)
		if err == nil {
			log.WithFields(log.Fields{
				"path":     s.Path,
				"kvconfig": s.Kvconfig,
			}).Info("succeeded in opening index")
		} else {
			log.WithFields(log.Fields{
				"path":     s.Path,
				"kvconfig": s.Kvconfig,
				"err":      err,
			}).Error("failed to open index")
		}
	}

	return err
}

func (s *BlastService) CloseIndex() error {
	err := s.Index.Close()
	if err == nil {
		log.WithFields(log.Fields{}).Info("succeeded in closing index")
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to close index")
	}

	return err
}

func (s *BlastService) GetIndex(ctx context.Context, req *proto.GetIndexRequest) (*proto.GetIndexResponse, error) {
	protoGetIndexResponse := &proto.GetIndexResponse{
		IndexPath: s.Path,
	}

	if req.IncludeIndexMapping {
		indexMapping, err := proto.MarshalAny(s.IndexMapping)
		if err != nil {
			return protoGetIndexResponse, err
		}
		protoGetIndexResponse.IndexMapping = &indexMapping
	}

	if req.IncludeIndexType {
		protoGetIndexResponse.IndexType = s.IndexType
	}

	if req.IncludeKvstore {
		protoGetIndexResponse.Kvstore = s.Kvstore
	}

	if req.IncludeKvconfig {
		kvconfig, err := proto.MarshalAny(s.Kvconfig)
		if err != nil {
			return protoGetIndexResponse, err
		}
		protoGetIndexResponse.Kvconfig = &kvconfig
	}

	return protoGetIndexResponse, nil
}

func (s *BlastService) PutDocument(ctx context.Context, req *proto.PutDocumentRequest) (*proto.PutDocumentResponse, error) {
	putCount := int32(0)
	fields, err := proto.UnmarshalAny(req.Document.Fields)
	if err == nil {
		log.WithFields(log.Fields{
			"id": req.Document.Id,
		}).Debug("succeeded in creating document")

		err = s.Index.Index(req.Document.Id, fields)
		if err == nil {
			putCount = 1

			log.WithFields(log.Fields{
				"id": req.Document.Id,
			}).Info("succeeded in putting document")
		} else {
			log.WithFields(log.Fields{
				"id":  req.Document.Id,
				"err": err,
			}).Error("failed to put document")
		}
	} else {
		log.WithFields(log.Fields{
			"id":  req.Document.Id,
			"err": err,
		}).Error("failed to put document")
	}

	return &proto.PutDocumentResponse{
		PutCount: putCount,
	}, err
}

func (s *BlastService) GetDocument(ctx context.Context, req *proto.GetDocumentRequest) (*proto.GetDocumentResponse, error) {
	fields := make(map[string]interface{})
	if doc, err := s.Index.Document(req.Id); err == nil {
		if doc != nil {
			log.WithFields(log.Fields{
				"id": req.Id,
			}).Info("succeeded in getting document")

			for _, field := range doc.Fields {
				var value interface{}

				switch field := field.(type) {
				case *document.TextField:
					value = string(field.Value())
				case *document.NumericField:
					numValue, err := field.Number()
					if err == nil {
						value = numValue
					}
				case *document.DateTimeField:
					dateValue, err := field.DateTime()
					if err == nil {
						dateValue.Format(time.RFC3339Nano)
						value = dateValue
					}
				}

				existedField, existed := fields[field.Name()]
				if existed {
					switch existedField := existedField.(type) {
					case []interface{}:
						fields[field.Name()] = append(existedField, value)
					case interface{}:
						arr := make([]interface{}, 2)
						arr[0] = existedField
						arr[1] = value
						fields[field.Name()] = arr
					}
				} else {
					fields[field.Name()] = value
				}
			}
		} else {
			log.WithFields(log.Fields{
				"id": req.Id,
			}).Info("document does not exist")
		}
	} else {
		log.WithFields(log.Fields{
			"id":  req.Id,
			"err": err,
		}).Error("failed to get document")

		return &proto.GetDocumentResponse{}, err
	}

	fieldsAny, err := proto.MarshalAny(fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":  req.Id,
			"err": err,
		}).Error("failed to get document")
	}

	document := proto.Document{
		Id:     req.Id,
		Fields: &fieldsAny,
	}

	return &proto.GetDocumentResponse{
		Document: &document,
	}, err
}

func (s *BlastService) DeleteDocument(ctx context.Context, req *proto.DeleteDocumentRequest) (*proto.DeleteDocumentResponse, error) {
	deleteCount := int32(0)
	err := s.Index.Delete(req.Id)
	if err == nil {
		deleteCount = 1
		log.WithFields(log.Fields{
			"id": req.Id,
		}).Info("succeeded in deleting document")
	} else {
		log.WithFields(log.Fields{
			"id":  req.Id,
			"err": err,
		}).Error("failed to delete document")
	}

	return &proto.DeleteDocumentResponse{
		DeleteCount: deleteCount,
	}, err
}

func (s *BlastService) Bulk(ctx context.Context, req *proto.BulkRequest) (*proto.BulkResponse, error) {
	var (
		batchCount    int32
		putCount      int32
		putErrorCount int32
		deleteCount   int32
	)

	batch := s.Index.NewBatch()

	for num, request := range req.Requests {
		switch request.Method {
		case "put":
			fields, err := proto.UnmarshalAny(request.Document.Fields)
			if err != nil {
				log.WithFields(log.Fields{
					"num":     num,
					"request": request,
				}).Warn("unexpected fields in request")

				continue
			}

			err = batch.Index(request.Document.Id, fields)
			if err == nil {
				log.WithFields(log.Fields{
					"num":     num,
					"request": request,
				}).Info("succeeded in putting document")

				putCount++
				batchCount++
			} else {
				log.WithFields(log.Fields{
					"num":     num,
					"request": request,
					"err":     err,
				}).Warn("failed to put document")

				putErrorCount++
			}
		case "delete":
			batch.Delete(request.Document.Id)

			log.WithFields(log.Fields{
				"num":     num,
				"request": request,
			}).Info("succeeded in deleting document")

			deleteCount++
			batchCount++
		default:
			log.WithFields(log.Fields{
				"num":     num,
				"request": request,
			}).Warn("unexpected method")

			continue
		}

		if batchCount%req.BatchSize == 0 {
			err := s.Index.Batch(batch)
			if err == nil {
				log.WithFields(log.Fields{
					"count": batch.Size(),
				}).Info("succeeded in indexing documents in bulk")
			} else {
				log.WithFields(log.Fields{
					"count": batch.Size(),
				}).Warn("failed to index  documents in bulk")
			}

			batch = s.Index.NewBatch()
		}
	}

	if batch.Size() > 0 {
		err := s.Index.Batch(batch)
		if err == nil {
			log.WithFields(log.Fields{
				"count": batch.Size(),
			}).Info("succeeded in indexing documents in bulk")
		} else {
			log.WithFields(log.Fields{
				"count": batch.Size(),
			}).Warn("failed to index  documents in bulk")
		}
	}

	return &proto.BulkResponse{
		PutCount:      putCount,
		PutErrorCount: putErrorCount,
		DeleteCount:   deleteCount,
	}, nil
}

func (s *BlastService) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	searchRequest, err := proto.UnmarshalAny(req.SearchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create search request")
		return &proto.SearchResponse{}, err
	}

	searchResult, err := s.Index.Search(searchRequest.(*bleve.SearchRequest))
	if err == nil {
		log.WithFields(log.Fields{}).Info("succeeded in searching documents")
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to search documents")

		return &proto.SearchResponse{}, err
	}

	searchResultAny, err := proto.MarshalAny(searchResult)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create search result")
	}

	return &proto.SearchResponse{
		SearchResult: &searchResultAny,
	}, err
}
