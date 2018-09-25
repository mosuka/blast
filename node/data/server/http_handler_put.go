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
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
)

type PutHandler struct {
	logger *log.Logger
	client *client.GRPCClient
}

func NewPutHandler(logger *log.Logger, client *client.GRPCClient) *PutHandler {
	return &PutHandler{
		logger: logger,
		client: client,
	}
}

func (h *PutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	fieldsBytes, err := ioutil.ReadAll(r.Body)
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

	// Check field length
	if len(fieldsBytes) <= 0 {
		err := errors.New("fields argument must be set")
		status = http.StatusBadRequest
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	var fieldsMap map[string]interface{}
	if fieldsBytes != nil {
		err = json.Unmarshal(fieldsBytes, &fieldsMap)
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
			status = http.StatusBadRequest
			errContent, err := NewContent(err.Error())
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}
			WriteResponse(w, errContent, status, h.logger)
			return
		}
	}

	fieldsAny := &any.Any{}
	if err = protobuf.UnmarshalAny(fieldsMap, fieldsAny); err != nil {
		h.logger.Printf("[ERR] %v", err)
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	req := &protobuf.PutDocumentRequest{
		Id:     id,
		Fields: fieldsAny,
	}

	resp, err := h.client.PutDocument(req)
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

	content, err := json.MarshalIndent(resp, "", "  ")
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
