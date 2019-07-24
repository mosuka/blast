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
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mosuka/blast/indexutils"

	"github.com/blevesearch/bleve"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewRouter(grpcAddr string, logger *zap.Logger) (*blasthttp.Router, error) {
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
	logger *zap.Logger
}

func NewRootHandler(logger *zap.Logger) *RootHandler {
	return &RootHandler{
		logger: logger,
	}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	msgMap := map[string]interface{}{
		"version": version.Version,
		"status":  status,
	}

	content, err := blasthttp.NewJSONMessage(msgMap)
	if err != nil {
		h.logger.Error(err.Error())
	}

	blasthttp.WriteResponse(w, content, status, h.logger)
}

type GetHandler struct {
	client *grpc.Client
	logger *zap.Logger
}

func NewGetDocumentHandler(client *grpc.Client, logger *zap.Logger) *GetHandler {
	return &GetHandler{
		client: client,
		logger: logger,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	vars := mux.Vars(r)

	fields, err := h.client.GetDocument(vars["id"])
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			status = http.StatusNotFound
		default:
			status = http.StatusInternalServerError
		}

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	// map[string]interface{} -> bytes
	content, err = json.MarshalIndent(fields, "", "  ")
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	blasthttp.WriteResponse(w, content, status, h.logger)
}

type IndexHandler struct {
	client *grpc.Client
	logger *zap.Logger
}

func NewSetDocumentHandler(client *grpc.Client, logger *zap.Logger) *IndexHandler {
	return &IndexHandler{
		client: client,
		logger: logger,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	// create documents
	docs := make([]*indexutils.Document, 0)

	vars := mux.Vars(r)
	id := vars["id"]

	bulk := func(values []string) bool {
		for _, value := range values {
			if strings.ToLower(value) == "true" {
				return true
			}
		}
		return false
	}(r.URL.Query()["bulk"])

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	if id == "" {
		if bulk {
			s := strings.NewReader(string(bodyBytes))
			reader := bufio.NewReader(s)
			for {
				docBytes, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF || err == io.ErrClosedPipe {
						if len(docBytes) > 0 {
							doc, err := indexutils.NewDocumentFromBytes(docBytes)
							if err != nil {
								status = http.StatusBadRequest

								msgMap := map[string]interface{}{
									"message": err.Error(),
									"status":  status,
								}

								content, err = blasthttp.NewJSONMessage(msgMap)
								if err != nil {
									h.logger.Error(err.Error())
								}

								blasthttp.WriteResponse(w, content, status, h.logger)
								return
							}
							docs = append(docs, doc)
						}
						break
					}
					status = http.StatusBadRequest

					msgMap := map[string]interface{}{
						"message": err.Error(),
						"status":  status,
					}

					content, err = blasthttp.NewJSONMessage(msgMap)
					if err != nil {
						h.logger.Error(err.Error())
					}

					blasthttp.WriteResponse(w, content, status, h.logger)
					return
				}

				if len(docBytes) > 0 {
					doc, err := indexutils.NewDocumentFromBytes(docBytes)
					if err != nil {
						status = http.StatusBadRequest

						msgMap := map[string]interface{}{
							"message": err.Error(),
							"status":  status,
						}

						content, err = blasthttp.NewJSONMessage(msgMap)
						if err != nil {
							h.logger.Error(err.Error())
						}

						blasthttp.WriteResponse(w, content, status, h.logger)
						return
					}
					docs = append(docs, doc)
				}
			}
		} else {
			doc, err := indexutils.NewDocumentFromBytes(bodyBytes)
			if err != nil {
				status = http.StatusBadRequest

				msgMap := map[string]interface{}{
					"message": err.Error(),
					"status":  status,
				}

				content, err = blasthttp.NewJSONMessage(msgMap)
				if err != nil {
					h.logger.Error(err.Error())
				}

				blasthttp.WriteResponse(w, content, status, h.logger)
				return
			}
			docs = append(docs, doc)
		}
	} else {
		var fields map[string]interface{}
		err = json.Unmarshal(bodyBytes, &fields)
		if err != nil {
			status = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  status,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Error(err.Error())
			}

			blasthttp.WriteResponse(w, content, status, h.logger)
			return
		}

		doc, err := indexutils.NewDocument(id, fields)
		if err != nil {
			status = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  status,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Error(err.Error())
			}

			blasthttp.WriteResponse(w, content, status, h.logger)
			return
		}

		docs = append(docs, doc)
	}

	// index documents in bulk
	count, err := h.client.IndexDocument(docs)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	// create JSON content
	msgMap := map[string]interface{}{
		"count": count,
	}
	content, err = json.MarshalIndent(msgMap, "", "  ")
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	blasthttp.WriteResponse(w, content, status, h.logger)
}

type DeleteHandler struct {
	client *grpc.Client
	logger *zap.Logger
}

func NewDeleteDocumentHandler(client *grpc.Client, logger *zap.Logger) *DeleteHandler {
	return &DeleteHandler{
		client: client,
		logger: logger,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	// create documents
	ids := make([]string, 0)

	vars := mux.Vars(r)
	id := vars["id"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	if id == "" {
		s := strings.NewReader(string(bodyBytes))
		reader := bufio.NewReader(s)
		for {
			docId, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					if docId == "" {
						ids = append(ids, docId)
					}
					break
				}
				status = http.StatusBadRequest

				msgMap := map[string]interface{}{
					"message": err.Error(),
					"status":  status,
				}

				content, err = blasthttp.NewJSONMessage(msgMap)
				if err != nil {
					h.logger.Error(err.Error())
				}

				blasthttp.WriteResponse(w, content, status, h.logger)
				return
			}

			if docId == "" {
				ids = append(ids, docId)
			}
		}
	} else {
		// Deleting a document
		ids = append(ids, id)
	}

	// delete documents in bulk
	count, err := h.client.DeleteDocument(ids)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	// create JSON content
	msgMap := map[string]interface{}{
		"count": count,
	}
	content, err = json.MarshalIndent(msgMap, "", "  ")
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}
}

type SearchHandler struct {
	client *grpc.Client
	logger *zap.Logger
}

func NewSearchHandler(client *grpc.Client, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		client: client,
		logger: logger,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	searchRequestBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	// []byte -> bleve.SearchRequest
	searchRequest := bleve.NewSearchRequest(nil)
	if len(searchRequestBytes) > 0 {
		err := json.Unmarshal(searchRequestBytes, searchRequest)
		if err != nil {
			status = http.StatusBadRequest

			msgMap := map[string]interface{}{
				"message": err.Error(),
				"status":  status,
			}

			content, err = blasthttp.NewJSONMessage(msgMap)
			if err != nil {
				h.logger.Error(err.Error())
			}

			blasthttp.WriteResponse(w, content, status, h.logger)
			return
		}
	}

	searchResult, err := h.client.Search(searchRequest)
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	content, err = json.MarshalIndent(&searchResult, "", "  ")
	if err != nil {
		status = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": err.Error(),
			"status":  status,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Error(err.Error())
		}

		blasthttp.WriteResponse(w, content, status, h.logger)
		return
	}

	blasthttp.WriteResponse(w, content, status, h.logger)
}
