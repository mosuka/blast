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
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
)

type BulkHandler struct {
	logger *log.Logger
	client *client.GRPCClient
}

func NewBulkHandler(logger *log.Logger, client *client.GRPCClient) *BulkHandler {
	return &BulkHandler{
		logger: logger,
		client: client,
	}
}

func (h *BulkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK

	defer HTTPMetrics(start, status, w, r, h.logger)

	var err error

	batchSize := int64(1000)
	batchSizeStr := r.URL.Query().Get("batch-size")
	if batchSizeStr != "" {
		batchSize, err = strconv.ParseInt(batchSizeStr, 10, 32)
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

	updatesBytes, err := ioutil.ReadAll(r.Body)
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

	// Check bulk request length
	if len(updatesBytes) <= 0 {
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

	updateMapSlice := make([]map[string]interface{}, 0)
	err = json.Unmarshal(updatesBytes, &updateMapSlice)
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

	updates := make([]*protobuf.BulkUpdateRequest_Update, 0)
	for _, updateMap := range updateMapSlice {
		update := &protobuf.BulkUpdateRequest_Update{}

		update.Command = protobuf.BulkUpdateRequest_Update_Command(protobuf.BulkUpdateRequest_Update_Command_value[updateMap["command"].(string)])

		documentMap, exist := updateMap["document"].(map[string]interface{})
		if !exist {
			err := errors.New("document does not exist")
			h.logger.Printf("[ERR] %v", err)
			status = http.StatusBadRequest
			errContent, err := NewContent(err.Error())
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}
			WriteResponse(w, errContent, status, h.logger)
			return
		}

		document := &protobuf.Document{}

		documentId, exist := documentMap["id"].(string)
		if !exist {
			err := errors.New("document id does not exist")
			h.logger.Printf("[ERR] %v", err)
			status = http.StatusBadRequest
			errContent, err := NewContent(err.Error())
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}
			WriteResponse(w, errContent, status, h.logger)
			return
		}
		document.Id = documentId

		fieldsMap, exist := documentMap["fields"].(map[string]interface{})
		if exist {
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
			document.Fields = fieldsAny
		}

		update.Document = document

		updates = append(updates, update)
	}

	req := &protobuf.BulkUpdateRequest{
		Updates:   updates,
		BatchSize: int32(batchSize),
	}

	resp, err := h.client.BulkUpdate(req)
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
