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
	IndexPath    string
	IndexMapping *mapping.IndexMappingImpl
	IndexType    string
	Kvstore      string
	Kvconfig     map[string]interface{}
	Index        bleve.Index
}

func NewBlastService(indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}) *BlastService {
	return &BlastService{
		IndexPath:    indexPath,
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
		Index:        nil,
	}
}

func (s *BlastService) OpenIndex() error {
	_, err := os.Stat(s.IndexPath)
	if os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"indexPath": s.IndexPath,
		}).Info("index does not exist")

		s.Index, err = bleve.NewUsing(s.IndexPath, s.IndexMapping, s.IndexType, s.Kvstore, s.Kvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"indexPath":    s.IndexPath,
				"indexMapping": s.IndexMapping,
				"indexType":    s.IndexType,
				"kvstore":      s.Kvstore,
				"kvconfig":     s.Kvconfig,
				"err":          err,
			}).Error("failed to create index")

			return err
		}

		log.WithFields(log.Fields{
			"indexPath":    s.IndexPath,
			"indexMapping": s.IndexMapping,
			"indexType":    s.IndexType,
			"kvstore":      s.Kvstore,
			"kvconfig":     s.Kvconfig,
		}).Info("succeeded in creating index")
	} else {
		log.WithFields(log.Fields{
			"indexPath": s.IndexPath,
		}).Info("index already exists")

		s.Index, err = bleve.OpenUsing(s.IndexPath, s.Kvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"indexPath": s.IndexPath,
				"kvconfig":  s.Kvconfig,
				"err":       err,
			}).Error("failed to open index")

			return err
		}

		log.WithFields(log.Fields{
			"indexPath": s.IndexPath,
			"kvconfig":  s.Kvconfig,
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

func (s *BlastService) GetIndexInfo(ctx context.Context, req *proto.GetIndexInfoRequest) (*proto.GetIndexInfoResponse, error) {
	protoGetIndexResponse := &proto.GetIndexInfoResponse{}

	if req.IndexPath {
		protoGetIndexResponse.IndexPath = s.IndexPath
	}

	if req.IndexMapping {
		indexMappingAny, err := proto.MarshalAny(s.IndexMapping)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"succeeded": false,
			}).Error("failed to marshal index mapping to Any type")

			return &proto.GetIndexInfoResponse{
				Succeeded: false,
				Message:   "failed to marshal index mapping to Any type",
			}, err
		}
		protoGetIndexResponse.IndexMapping = &indexMappingAny
	}

	if req.IndexType {
		protoGetIndexResponse.IndexType = s.IndexType
	}

	if req.Kvstore {
		protoGetIndexResponse.Kvstore = s.Kvstore
	}

	if req.Kvconfig {
		kvconfigAny, err := proto.MarshalAny(s.Kvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"succeeded": false,
			}).Error("failed to marshal kvconfig to Any type")

			return &proto.GetIndexInfoResponse{
				Succeeded: false,
				Message:   "failed to marshal kvconfig to Any type",
			}, err
		}
		protoGetIndexResponse.Kvconfig = &kvconfigAny
	}

	protoGetIndexResponse.Succeeded = true
	protoGetIndexResponse.Message = "succeeded in get an index information"

	return protoGetIndexResponse, nil
}

func (s *BlastService) PutDocument(ctx context.Context, req *proto.PutDocumentRequest) (*proto.PutDocumentResponse, error) {
	fields, err := proto.UnmarshalAny(req.Fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":        req.Id,
			"fields":    req.Fields,
			"err":       err,
			"succeeded": false,
		}).Error("failed to unmarshal fields")

		return &proto.PutDocumentResponse{
			Id:        req.Id,
			Fields:    req.Fields,
			Succeeded: false,
			Message:   "failed to unmarshal fields",
		}, err
	}

	err = s.Index.Index(req.Id, fields)
	if err != nil {
		log.WithFields(log.Fields{
			"id":        req.Id,
			"fields":    fields,
			"err":       err,
			"succeeded": false,
		}).Error("failed to put a document")

		return &proto.PutDocumentResponse{
			Id:        req.Id,
			Fields:    req.Fields,
			Succeeded: false,
			Message:   "failed to put a document",
		}, err
	}

	log.WithFields(log.Fields{
		"id":        req.Id,
		"fields":    fields,
		"succeeded": true,
	}).Info("succeeded in put a document")

	return &proto.PutDocumentResponse{
		Id:        req.Id,
		Fields:    req.Fields,
		Succeeded: true,
		Message:   "succeeded in put a document",
	}, nil
}

func (s *BlastService) GetDocument(ctx context.Context, req *proto.GetDocumentRequest) (*proto.GetDocumentResponse, error) {
	doc, err := s.Index.Document(req.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"id":        req.Id,
			"err":       err,
			"succeeded": false,
		}).Error("failed to get a document")

		return &proto.GetDocumentResponse{
			Id:        req.Id,
			Succeeded: false,
			Message:   "failed to get a document",
		}, err
	}

	if doc == nil {
		log.WithFields(log.Fields{
			"id":        req.Id,
			"succeeded": true,
		}).Info("document does not exist")

		return &proto.GetDocumentResponse{
			Id:        req.Id,
			Succeeded: true,
			Message:   "document does not exist",
		}, nil
	}

	fields := make(map[string]interface{})
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
			"id":        req.Id,
			"fields":    fields,
			"err":       err,
			"succeeded": false,
		}).Error("failed to marshal fields")

		return &proto.GetDocumentResponse{
			Id:        req.Id,
			Fields:    &fieldsAny,
			Succeeded: false,
			Message:   "failed to marshal fields",
		}, err
	}

	log.WithFields(log.Fields{
		"id":        req.Id,
		"fields":    fields,
		"succeeded": true,
	}).Info("succeeded in get a document")

	return &proto.GetDocumentResponse{
		Id:        req.Id,
		Fields:    &fieldsAny,
		Succeeded: true,
		Message:   "succeeded in get a document",
	}, err
}

func (s *BlastService) DeleteDocument(ctx context.Context, req *proto.DeleteDocumentRequest) (*proto.DeleteDocumentResponse, error) {
	err := s.Index.Delete(req.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"id":        req.Id,
			"err":       err,
			"succeeded": false,
		}).Error("failed to delete a document")

		return &proto.DeleteDocumentResponse{
			Id:        req.Id,
			Succeeded: false,
			Message:   "failed to delete a document",
		}, err
	}

	log.WithFields(log.Fields{
		"id":        req.Id,
		"succeeded": true,
	}).Info("succeeded in delete a document")

	return &proto.DeleteDocumentResponse{
		Id:        req.Id,
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
					"num":    num,
					"id":     updateRequest.Document.Id,
					"fields": fields,
					"err":    err,
				}).Warn("failed to put a document")

				putErrorCount++
			}

			log.WithFields(log.Fields{
				"num":    num,
				"id":     updateRequest.Document.Id,
				"fields": fields,
			}).Debug("succeeded in put a document")

			putCount++
			processedCount++
		case "delete":
			batch.Delete(updateRequest.Document.Id)

			log.WithFields(log.Fields{
				"num": num,
				"id":  updateRequest.Document.Id,
			}).Debug("succeeded in delete a document")

			deleteCount++
			processedCount++
		default:
			log.WithFields(log.Fields{
				"num":    num,
				"method": updateRequest.Method,
			}).Warn("unknown method")

			continue
		}

		if processedCount%req.BatchSize == 0 {
			err := s.Index.Batch(batch)
			if err == nil {
				log.WithFields(log.Fields{
					"size": batch.Size(),
				}).Debug("succeeded in put documents in bulk")
			} else {
				log.WithFields(log.Fields{
					"size": batch.Size(),
				}).Warn("failed to put documents in bulk")
			}

			batch = s.Index.NewBatch()
		}
	}

	if batch.Size() > 0 {
		err := s.Index.Batch(batch)
		if err == nil {
			log.WithFields(log.Fields{
				"size": batch.Size(),
			}).Debug("succeeded in put documents in bulk")
		} else {
			log.WithFields(log.Fields{
				"size": batch.Size(),
			}).Warn("failed to put documents in bulk")
		}
	}

	log.WithFields(log.Fields{
		"putCount":      putCount,
		"putErrorCount": putErrorCount,
		"deleteCount":   deleteCount,
		"succeeded":     true,
	}).Info("succeeded in put documents in bulk")

	return &proto.BulkResponse{
		PutCount:      putCount,
		PutErrorCount: putErrorCount,
		DeleteCount:   deleteCount,
		Succeeded:     true,
		Message:       "succeeded in put documents in bulk",
	}, nil
}

func (s *BlastService) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	searchRequest, err := proto.UnmarshalAny(req.SearchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err,
			"succeeded": false,
		}).Error("failed to unmarshal search request to Any type")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   "failed to unmarshal search request to Any type",
		}, err
	}

	searchResult, err := s.Index.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err,
			"succeeded": false,
		}).Error("failed to search documents")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   "failed to search documents",
		}, err
	}

	searchResultAny, err := proto.MarshalAny(searchResult)
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err,
			"succeeded": false,
		}).Error("failed to marshal fields to Any type")

		return &proto.SearchResponse{
			Succeeded: false,
			Message:   "failed to marshal fields to Any type",
		}, err
	}

	log.WithFields(log.Fields{
		"searchResult": searchResult,
		"succeeded":    true,
	}).Info("succeeded in searching documents")

	return &proto.SearchResponse{
		SearchResult: &searchResultAny,
		Succeeded:    true,
		Message:      "succeeded in search documents",
	}, nil
}
