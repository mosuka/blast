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

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/errors"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Router struct {
	mux.Router

	GRPCClient *GRPCClient
	logger     *zap.Logger
}

func NewRouter(grpcAddr string, logger *zap.Logger) (*Router, error) {
	grpcClient, err := NewGRPCClient(grpcAddr)
	if err != nil {
		return nil, err
	}

	router := &Router{
		GRPCClient: grpcClient,
		logger:     logger,
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

func (r *Router) Close() error {
	r.GRPCClient.Cancel()

	err := r.GRPCClient.Close()
	if err != nil {
		return err
	}

	return nil
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
	client *GRPCClient
	logger *zap.Logger
}

func NewGetDocumentHandler(client *GRPCClient, logger *zap.Logger) *GetHandler {
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

	doc, err := h.client.GetDocument(vars["id"])
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

	content, err = index.MarshalDocument(doc)
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
	client *GRPCClient
	logger *zap.Logger
}

func NewSetDocumentHandler(client *GRPCClient, logger *zap.Logger) *IndexHandler {
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
	docs := make([]*index.Document, 0)

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
							var doc *index.Document
							err = proto.Unmarshal(bodyBytes, doc)
							//doc, err := indexutils.NewDocumentFromBytes(docBytes)
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
					doc := &index.Document{}
					err = index.UnmarshalDocument(docBytes, doc)
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
			doc := &index.Document{}
			err = index.UnmarshalDocument(bodyBytes, doc)
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
		var fieldsMap map[string]interface{}
		err := json.Unmarshal([]byte(bodyBytes), &fieldsMap)
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
		docMap := map[string]interface{}{
			"id":     id,
			"fields": fieldsMap,
		}
		docBytes, err := json.Marshal(docMap)
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
		doc := &index.Document{}
		err = index.UnmarshalDocument(docBytes, doc)
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
	client *GRPCClient
	logger *zap.Logger
}

func NewDeleteDocumentHandler(client *GRPCClient, logger *zap.Logger) *DeleteHandler {
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
	client *GRPCClient
	logger *zap.Logger
}

func NewSearchHandler(client *GRPCClient, logger *zap.Logger) *SearchHandler {
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
