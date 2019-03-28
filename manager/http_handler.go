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
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/gorilla/mux"
	blasterrors "github.com/mosuka/blast/errors"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
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

	key := "/" + vars["path"]

	kvp := &management.KeyValuePair{
		Key: key,
	}

	retKVP, err := h.client.Get(kvp)
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

	// Any -> map[string]interface{}
	valueInstance, err := protobuf.MarshalAny(retKVP.Value)
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
	if valueInstance == nil {
		httpStatus = http.StatusInternalServerError

		msgMap := map[string]interface{}{
			"message": errors.New("nil"),
			"status":  httpStatus,
		}

		content, err = blasthttp.NewJSONMessage(msgMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}

		return
	}
	valueMap := valueInstance.(*map[string]interface{})

	// map[string]interface -> []byte
	content, err = json.MarshalIndent(valueMap, "", "  ")
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
	client *GRPCClient
	logger *log.Logger
}

func NewPutHandler(client *GRPCClient, logger *log.Logger) *PutHandler {
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

	key := "/" + vars["path"]

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
	var valueMap map[string]interface{}
	err = json.Unmarshal(bodyBytes, &valueMap)
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

	// map[string]interface{} -> Any
	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny(valueMap, valueAny)
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

	kvp := &management.KeyValuePair{
		Key:   key,
		Value: valueAny,
	}

	err = h.client.Set(kvp)
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

	vars := mux.Vars(r)

	key := "/" + vars["path"]

	kvp := &management.KeyValuePair{
		Key: key,
	}

	err := h.client.Delete(kvp)
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
