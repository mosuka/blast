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
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/errors"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	pbindex "github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/version"
)

type RootHandler struct {
	logger *log.Logger
}

func NewRootHandler(logger *log.Logger) *RootHandler {
	return &RootHandler{
		logger: logger,
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, status, h.logger)
		blasthttp.RecordMetrics(start, status, w, r, h.logger)
	}()

	msgMap := map[string]interface{}{
		"version": version.Version,
		"status":  status,
	}

	content, err := blasthttp.NewJSONMessage(msgMap)
	if err != nil {
		h.logger.Printf("[ERR] %v", err)
	}
}

type GetHandler struct {
	client *GRPCClient
	logger *log.Logger
}

func NewGetHandler(client *GRPCClient, logger *log.Logger) *GetHandler {
	return &GetHandler{
		client: client,
		logger: logger,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	vars := mux.Vars(r)

	doc := &index.Document{
		Id: vars["id"],
	}

	doc, err := h.client.Get(doc)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			httpStatus = http.StatusNotFound
		default:
			httpStatus = http.StatusInternalServerError
		}

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	// Any -> map[string]interface{}
	var fieldsMap *map[string]interface{}
	fieldsInstance, err := protobuf.MarshalAny(doc.Fields)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
	if fieldsInstance == nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
	fieldsMap = fieldsInstance.(*map[string]interface{})

	// map[string]interface{} -> bytes
	content, err = json.MarshalIndent(fieldsMap, "", "  ")
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
}

type IndexHandler struct {
	client *GRPCClient
	logger *log.Logger
}

func NewIndexHandler(client *GRPCClient, logger *log.Logger) *IndexHandler {
	return &IndexHandler{
		client: client,
		logger: logger,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	// create documents
	docs := make([]*index.Document, 0)

	vars := mux.Vars(r)
	id := vars["id"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	if id == "" {
		// Indexing documents in bulk
		var docMaps []map[string]interface{}
		err := json.Unmarshal(bodyBytes, &docMaps)
		if err != nil {
			httpStatus = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  httpStatus,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}

			return
		}

		for _, docMap := range docMaps {
			fieldsAny := &any.Any{}
			err = protobuf.UnmarshalAny(docMap["fields"], fieldsAny)
			if err != nil {
				httpStatus = http.StatusBadRequest

				msgMap := map[string]interface{}{
					"message": err.Error(),
					"status":  httpStatus,
				}

				content, err = blasthttp.NewJSONMessage(msgMap)
				if err != nil {
					h.logger.Printf("[ERR] %v", err)
				}

				return
			}

			doc := &pbindex.Document{
				Id:     docMap["id"].(string),
				Fields: fieldsAny,
			}

			docs = append(docs, doc)
		}
	} else {
		// Indexing a document
		var fieldsMap map[string]interface{}
		err := json.Unmarshal(bodyBytes, &fieldsMap)
		if err != nil {
			httpStatus = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  httpStatus,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}

			return
		}

		fieldsAny := &any.Any{}
		err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
		if err != nil {
			httpStatus = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  httpStatus,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}

			return
		}

		doc := &pbindex.Document{
			Id:     id,
			Fields: fieldsAny,
		}

		docs = append(docs, doc)
	}

	// index documents in bulk
	result, err := h.client.Index(docs)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	content, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
}

type DeleteHandler struct {
	client *GRPCClient
	logger *log.Logger
}

func NewDeleteHandler(client *GRPCClient, logger *log.Logger) *DeleteHandler {
	return &DeleteHandler{
		client: client,
		logger: logger,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	// create documents
	docs := make([]*index.Document, 0)

	vars := mux.Vars(r)
	id := vars["id"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	if id == "" {
		// Deleting documents in bulk
		var docMaps []map[string]interface{}
		err := json.Unmarshal(bodyBytes, &docMaps)
		if err != nil {
			httpStatus = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  httpStatus,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}

			return
		}

		for _, docMap := range docMaps {
			doc := &pbindex.Document{
				Id: docMap["id"].(string),
			}

			docs = append(docs, doc)
		}
	} else {
		// Deleting a document
		doc := &pbindex.Document{
			Id: id,
		}

		docs = append(docs, doc)
	}

	// delete documents in bulk
	result, err := h.client.Delete(docs)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	content, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
}

type SearchHandler struct {
	client *GRPCClient
	logger *log.Logger
}

func NewSearchHandler(client *GRPCClient, logger *log.Logger) *SearchHandler {
	return &SearchHandler{
		client: client,
		logger: logger,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	searchRequestBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	// []byte -> bleve.SearchRequest
	searchRequest := bleve.NewSearchRequest(nil)
	if len(searchRequestBytes) > 0 {
		err := json.Unmarshal(searchRequestBytes, searchRequest)
		if err != nil {
			httpStatus = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  httpStatus,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}

			return
		}
	}

	searchResult, err := h.client.Search(searchRequest)
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

	content, err = json.MarshalIndent(&searchResult, "", "  ")
	if err != nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}

}
