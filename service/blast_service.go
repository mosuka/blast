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
	"fmt"
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
		if err != nil {
			log.WithFields(log.Fields{
				"path":         s.Path,
				"indexMapping": s.IndexMapping,
				"indexType":    s.IndexType,
				"kvstore":      s.Kvstore,
				"kvconfig":     s.Kvconfig,
				"err":          err,
			}).Error("failed to create index")

			return err
		}

		log.WithFields(log.Fields{
			"path":         s.Path,
			"indexMapping": s.IndexMapping,
			"indexType":    s.IndexType,
			"kvstore":      s.Kvstore,
			"kvconfig":     s.Kvconfig,
		}).Info("succeeded in creating index")
	} else {
		log.WithFields(log.Fields{
			"path": s.Path,
		}).Info("index already exists")

		s.Index, err = bleve.OpenUsing(s.Path, s.Kvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"path":     s.Path,
				"kvconfig": s.Kvconfig,
				"err":      err,
			}).Error("failed to open index")

			return err
		}

		log.WithFields(log.Fields{
			"path":     s.Path,
			"kvconfig": s.Kvconfig,
		}).Info("succeeded in opening index")
	}

	return nil
}

func (s *BlastService) CloseIndex() error {
	err := s.Index.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to close index")

		return err
	}

	log.WithFields(log.Fields{}).Info("succeeded in closing index")

	return nil
}

func (s *BlastService) GetIndex(ctx context.Context, req *proto.GetIndexRequest) (*proto.GetIndexResponse, error) {
	protoGetIndexResponse := &proto.GetIndexResponse{
		IndexPath: s.Path,
	}

	if req.IncludeIndexMapping {
		err := protoGetIndexResponse.SetIndexMappingActual(s.IndexMapping)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to marshal index mapping to Any type")

			return &proto.GetIndexResponse{
				Succeeded: false,
				Message:   fmt.Sprintf("failed to marshal index mapping to Any type : %s", err.Error()),
			}, err
		}
	}

	if req.IncludeIndexType {
		protoGetIndexResponse.IndexType = s.IndexType
	}

	if req.IncludeKvstore {
		protoGetIndexResponse.Kvstore = s.Kvstore
	}

	if req.IncludeKvconfig {
		err := protoGetIndexResponse.SetKvconfigActual(s.Kvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to marshal kvconfig to Any type")

			return &proto.GetIndexResponse{
				Succeeded: false,
				Message:   fmt.Sprintf("failed to marshal kvconfig to Any type : %s", err.Error()),
			}, err
		}
	}

	protoGetIndexResponse.Succeeded = true
	protoGetIndexResponse.Message = "succeeded in get an index information"

	return protoGetIndexResponse, nil
}

func (s *BlastService) PutDocument(ctx context.Context, req *proto.PutDocumentRequest) (*proto.PutDocumentResponse, error) {
	fields, err := proto.UnmarshalAny(req.Document.Fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":  req.Document.Id,
			"err": err,
		}).Error("failed to unmarshal fields to Any type")

		return &proto.PutDocumentResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to marshal fields to Any type : %s", err.Error()),
		}, err
	}

	log.WithFields(log.Fields{
		"id":     req.Document.Id,
		"fields": fields,
	}).Debug("fields creation succeeded")

	err = s.Index.Index(req.Document.Id, fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":     req.Document.Id,
			"fields": fields,
			"err":    err,
		}).Error("failed to put a document")

		return &proto.PutDocumentResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to put a document : %s", err.Error()),
		}, err
	}

	log.WithFields(log.Fields{
		"id": req.Document.Id,
	}).Info("succeeded in put a document")

	return &proto.PutDocumentResponse{
		Succeeded: true,
		Message:   "succeeded in put a document",
	}, nil
}

func (s *BlastService) GetDocument(ctx context.Context, req *proto.GetDocumentRequest) (*proto.GetDocumentResponse, error) {
	fields := make(map[string]interface{})

	doc, err := s.Index.Document(req.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"id":  req.Id,
			"err": err,
		}).Error("failed to get a document")

		return &proto.GetDocumentResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to get a document : %s", err.Error()),
		}, err
	}

	if doc == nil {
		log.WithFields(log.Fields{
			"id": req.Id,
		}).Info("document does not exist")

		return &proto.GetDocumentResponse{
			Succeeded: true,
			Message:   "document does not exist",
		}, nil
	}

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

	fieldsAny, err := proto.MarshalAny(fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":     req.Id,
			"fields": fields,
			"err":    err,
		}).Error("failed to marshal fields to Any type")

		return &proto.GetDocumentResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to marshal fields to Any type : %s", err.Error()),
		}, err
	}

	document := &proto.Document{
		Id:     req.Id,
		Fields: &fieldsAny,
	}

	log.WithFields(log.Fields{
		"id": req.Id,
	}).Info("succeeded in get a document")

	return &proto.GetDocumentResponse{
		Succeeded: true,
		Document:  document,
		Message:   "succeeded in get a document",
	}, err
}

func (s *BlastService) DeleteDocument(ctx context.Context, req *proto.DeleteDocumentRequest) (*proto.DeleteDocumentResponse, error) {
	err := s.Index.Delete(req.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"id":  req.Id,
			"err": err,
		}).Error("failed to delete a document")

		return &proto.DeleteDocumentResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to delete a document : %s", err.Error()),
		}, err
	}

	log.WithFields(log.Fields{
		"id": req.Id,
	}).Info("succeeded in delete a document")

	return &proto.DeleteDocumentResponse{
		Succeeded: true,
		Message:   "succeeded in delete a document",
	}, nil
}

func (s *BlastService) Bulk(ctx context.Context, req *proto.BulkRequest) (*proto.BulkResponse, error) {
	var (
		processedCount int32
		putCount       int32
		putErrorCount  int32
		deleteCount    int32
	)

	batch := s.Index.NewBatch()

	for num, updateRequest := range req.UpdateRequests {
		switch updateRequest.Method {
		case "put":
			fields, err := proto.UnmarshalAny(updateRequest.Document.Fields)
			if err != nil {
				log.WithFields(log.Fields{
					"num":           num,
					"updateRequest": updateRequest,
				}).Warn("failed to unmarshal fields to Any type")

				continue
			}

			err = batch.Index(updateRequest.Document.Id, fields)
			if err != nil {
				log.WithFields(log.Fields{
					"num":           num,
					"updateRequest": updateRequest,
					"err":           err,
				}).Warn("failed to put document")

				putErrorCount++
			}

			log.WithFields(log.Fields{
				"num":           num,
				"updateRequest": updateRequest,
			}).Debug("succeeded in put a document")

			putCount++
			processedCount++
		case "delete":
			batch.Delete(updateRequest.Document.Id)

			log.WithFields(log.Fields{
				"num":           num,
				"updateRequest": updateRequest,
			}).Debug("succeeded in delete a document")

			deleteCount++
			processedCount++
		default:
			log.WithFields(log.Fields{
				"num":           num,
				"updateRequest": updateRequest,
			}).Warn("unknown method")

			continue
		}

		if processedCount%req.BatchSize == 0 {
			err := s.Index.Batch(batch)
			if err == nil {
				log.WithFields(log.Fields{
					"count": batch.Size(),
				}).Debug("succeeded in put documents in bulk")
			} else {
				log.WithFields(log.Fields{
					"count": batch.Size(),
				}).Warn("failed to put documents in bulk")
			}

			batch = s.Index.NewBatch()
		}
	}

	if batch.Size() > 0 {
		err := s.Index.Batch(batch)
		if err == nil {
			log.WithFields(log.Fields{
				"count": batch.Size(),
			}).Debug("succeeded in put documents in bulk")
		} else {
			log.WithFields(log.Fields{
				"count": batch.Size(),
			}).Warn("failed to put documents in bulk")
		}
	}

	log.WithFields(log.Fields{
		"putCount":      putCount,
		"putErrorCount": putErrorCount,
		"deleteCount":   deleteCount,
	}).Info("succeeded in put documents in bulk")

	return &proto.BulkResponse{
		Succeeded:     true,
		Message:       "succeeded in put documents in bulk",
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
		}).Error("failed to unmarshal search request to Any type")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to unmarshal search request to Any type : %s", err.Error()),
		}, err
	}

	searchResult, err := s.Index.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to search documents")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to search documents : %s", err.Error()),
		}, err
	}

	searchResultAny, err := proto.MarshalAny(searchResult)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to marshal fields to Any type")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   fmt.Sprintf("failed to marshal fields to Any type : %s", err.Error()),
		}, err
	}

	log.WithFields(log.Fields{
		"hits":     searchResult.Hits,
		"took":     searchResult.Took,
		"total":    searchResult.Total,
		"status":   searchResult.Status,
		"requests": searchResult.Request,
		"facets":   searchResult.Facets,
		"maxScore": searchResult.MaxScore,
	}).Info("succeeded in searching documents")

	return &proto.SearchResponse{
		SearchResult: &searchResultAny,
		Succeeded:    true,
		Message:      "succeeded in search documents",
	}, nil
}
