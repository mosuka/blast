// Copyright (c) 2018 Minoru Osuka
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

package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
)

type GetHandler struct {
	logger *log.Logger
	client *client.GRPCClient
}

func NewGetHandler(logger *log.Logger, client *client.GRPCClient) *GetHandler {
	return &GetHandler{
		logger: logger,
		client: client,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	defer HTTPMetrics(start, status, w, r, h.logger)

	var err error

	vars := mux.Vars(r)

	id := vars["id"]
	if id == "" {
		err := errors.New("update requests argument must be set")
		h.logger.Printf("[ERR] %v", err)
		status = http.StatusBadRequest
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	req := &protobuf.GetDocumentRequest{
		Id: id,
	}

	resp, err := h.client.GetDocument(req)
	if err != nil {
		h.logger.Printf("[ERR] %v", err)
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	// Document does not exist
	if resp.Document == nil {
		status = http.StatusNotFound
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, make([]byte, 0), status, h.logger)
		return
	}

	fieldsInstance, err := protobuf.MarshalAny(resp.Document.Fields)
	if err != nil {
		h.logger.Printf("[ERR] %v", err)
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	document := struct {
		Id     string                 `json:"id,omitempty"`
		Fields map[string]interface{} `json:"fields,omitempty"`
	}{
		Id:     resp.Document.Id,
		Fields: *fieldsInstance.(*map[string]interface{}),
	}

	content, err := json.MarshalIndent(&document, "", "  ")
	if err != nil {
		h.logger.Printf("[ERR] %v", err)
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	WriteResponse(w, content, status, h.logger)

	return
}
