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

package dispatcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/blevesearch/bleve"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(grpcAddr string, logger *log.Logger) (*blasthttp.Router, error) {
	router, err := blasthttp.NewRouter(grpcAddr, logger)
	if err != nil {
		return nil, err
	}

	router.StrictSlash(true)

	router.Handle("/", NewRootHandler(logger)).Methods("GET")
	router.Handle("/documents", NewSetDocumentHandler(router.GRPCClient, logger)).Methods("PUT")
	router.Handle("/documents", NewDeleteDocumentHandler(router.GRPCClient, logger)).Methods("DELETE")
	router.Handle("/documents/{id}", NewGetDocumentHandler(router.GRPCClient, logger)).Methods("GET")
	router.Handle("/documents/{id}", NewSetDocumentHandler(router.GRPCClient, logger)).Methods("PUT")
	router.Handle("/documents/{id}", NewDeleteDocumentHandler(router.GRPCClient, logger)).Methods("DELETE")
	router.Handle("/search", NewSearchHandler(router.GRPCClient, logger)).Methods("POST")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	return router, nil
}

type RootHandler struct {
	logger *log.Logger
}

func NewRootHandler(logger *log.Logger) *RootHandler {
	return &RootHandler{
		logger: logger,
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, status, h.logger)
		//blasthttp.RecordMetrics(start, status, w, r, h.logger)
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
	client *grpc.Client
	logger *log.Logger
}

func NewGetDocumentHandler(client *grpc.Client, logger *log.Logger) *GetHandler {
	return &GetHandler{
		client: client,
		logger: logger,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		//blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	vars := mux.Vars(r)

	fields, err := h.client.GetDocument(vars["id"])
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

	// map[string]interface{} -> bytes
	content, err = json.MarshalIndent(fields, "", "  ")
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
	client *grpc.Client
	logger *log.Logger
}

func NewSetDocumentHandler(client *grpc.Client, logger *log.Logger) *IndexHandler {
	return &IndexHandler{
		client: client,
		logger: logger,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		//blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	// create documents
	docs := make([]map[string]interface{}, 0)

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
		err := json.Unmarshal(bodyBytes, &docs)
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
	} else {
		// Indexing a document
		var fields map[string]interface{}
		err := json.Unmarshal(bodyBytes, &fields)
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

		doc := map[string]interface{}{
			"id":     id,
			"fields": fields,
		}

		docs = append(docs, doc)
	}

	// index documents in bulk
	count, err := h.client.IndexDocument(docs)
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

	// create JSON content
	msgMap := map[string]interface{}{
		"count": count,
	}
	content, err = json.MarshalIndent(msgMap, "", "  ")
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
	client *grpc.Client
	logger *log.Logger
}

func NewDeleteDocumentHandler(client *grpc.Client, logger *log.Logger) *DeleteHandler {
	return &DeleteHandler{
		client: client,
		logger: logger,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		//blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	// create documents
	ids := make([]string, 0)

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
		err := json.Unmarshal(bodyBytes, &ids)
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
	} else {
		// Deleting a document
		ids = append(ids, id)
	}

	// delete documents in bulk
	count, err := h.client.DeleteDocument(ids)
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

	// create JSON content
	msgMap := map[string]interface{}{
		"count": count,
	}
	content, err = json.MarshalIndent(msgMap, "", "  ")
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
	client *grpc.Client
	logger *log.Logger
}

func NewSearchHandler(client *grpc.Client, logger *log.Logger) *SearchHandler {
	return &SearchHandler{
		client: client,
		logger: logger,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		//blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
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
