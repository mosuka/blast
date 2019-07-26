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
	"net/http"
	"time"

	"github.com/gorilla/mux"
	blasterrors "github.com/mosuka/blast/errors"
	blasthttp "github.com/mosuka/blast/http"
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
	router.Handle("/configs", NewPutHandler(router.GRPCClient, logger)).Methods("PUT")
	router.Handle("/configs", NewGetHandler(router.GRPCClient, logger)).Methods("GET")
	router.Handle("/configs", NewDeleteHandler(router.GRPCClient, logger)).Methods("DELETE")
	router.Handle("/configs/{path:.*}", NewPutHandler(router.GRPCClient, logger)).Methods("PUT")
	router.Handle("/configs/{path:.*}", NewGetHandler(router.GRPCClient, logger)).Methods("GET")
	router.Handle("/configs/{path:.*}", NewDeleteHandler(router.GRPCClient, logger)).Methods("DELETE")
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

func NewGetHandler(client *GRPCClient, logger *zap.Logger) *GetHandler {
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

	key := vars["path"]

	value, err := h.client.GetValue(key)
	if err != nil {
		switch err {
		case blasterrors.ErrNotFound:
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

	// interface{} -> []byte
	content, err = json.MarshalIndent(value, "", "  ")
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

type PutHandler struct {
	client *GRPCClient
	logger *zap.Logger
}

func NewPutHandler(client *GRPCClient, logger *zap.Logger) *PutHandler {
	return &PutHandler{
		client: client,
		logger: logger,
	}
}

func (h *PutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	content := make([]byte, 0)

	defer blasthttp.RecordMetrics(start, status, w, r)

	vars := mux.Vars(r)

	key := vars["path"]

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

	// string -> map[string]interface{}
	var value interface{}
	err = json.Unmarshal(bodyBytes, &value)
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

	err = h.client.SetValue(key, value)
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

func NewDeleteHandler(client *GRPCClient, logger *zap.Logger) *DeleteHandler {
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

	vars := mux.Vars(r)

	key := vars["path"]

	err := h.client.DeleteValue(key)
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
