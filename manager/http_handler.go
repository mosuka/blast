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

package manager

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	*mux.Router

	grpcClient *grpc.Client
}

func NewRouter(grpcAddr string, logger *log.Logger) (*Router, error) {
	grpcClient, err := grpc.NewClient(grpcAddr)
	if err != nil {
		return nil, err
	}

	router := &Router{
		grpcClient: grpcClient,
	}

	router.StrictSlash(true)

	router.Handle("/", NewRootHandler(logger)).Methods("GET")
	router.Handle("/configs", NewPutHandler(grpcClient, logger)).Methods("PUT")
	router.Handle("/configs", NewGetHandler(grpcClient, logger)).Methods("GET")
	router.Handle("/configs", NewDeleteHandler(grpcClient, logger)).Methods("DELETE")
	router.Handle("/configs/{path:.*}", NewPutHandler(grpcClient, logger)).Methods("PUT")
	router.Handle("/configs/{path:.*}", NewGetHandler(grpcClient, logger)).Methods("GET")
	router.Handle("/configs/{path:.*}", NewDeleteHandler(grpcClient, logger)).Methods("DELETE")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	return router, nil
}

func (r *Router) Close() error {
	r.grpcClient.Cancel()

	err := r.grpcClient.Close()
	if err != nil {
		return err
	}

	return nil
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
	client *grpc.Client
	logger *log.Logger
}

func NewGetHandler(client *grpc.Client, logger *log.Logger) *GetHandler {
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

	key := vars["path"]

	value, err := h.client.GetState(key)
	if err != nil {
		switch err {
		case blasterrors.ErrNotFound:
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

	// interface{} -> []byte
	content, err = json.MarshalIndent(value, "", "  ")
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

type PutHandler struct {
	client *grpc.Client
	logger *log.Logger
}

func NewPutHandler(client *grpc.Client, logger *log.Logger) *PutHandler {
	return &PutHandler{
		client: client,
		logger: logger,
	}
}

func (h *PutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	httpStatus := http.StatusOK
	content := make([]byte, 0)
	defer func() {
		blasthttp.WriteResponse(w, content, httpStatus, h.logger)
		blasthttp.RecordMetrics(start, httpStatus, w, r, h.logger)
	}()

	vars := mux.Vars(r)

	key := vars["path"]

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

	// string -> map[string]interface{}
	var value interface{}
	err = json.Unmarshal(bodyBytes, &value)
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

	err = h.client.SetState(key, value)
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

func NewDeleteHandler(client *grpc.Client, logger *log.Logger) *DeleteHandler {
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

	vars := mux.Vars(r)

	key := vars["path"]

	err := h.client.DeleteState(key)
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
