//  Copyright (c) 2018 Minoru Osuka
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

package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mosuka/blast/grpc/client"
	"github.com/mosuka/blast/http/metrics"
	"github.com/mosuka/blast/protobuf"
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
	var err error

	start := time.Now()
	status := http.StatusOK
	defer metrics.Metrics(start, status, w, r, h.logger)

	batchSize, err := strconv.ParseInt(r.URL.Query().Get("batch-size"), 10, 32)
	prettyPrint, err := strconv.ParseBool(r.URL.Query().Get("pretty-print"))

	updateRequestsBytes := make([]byte, 0)
	if updateRequestsBytes, err = ioutil.ReadAll(r.Body); err != nil {
		h.logger.Printf("[ERR] handler: Failed to read request body: %s", err.Error())
		status = http.StatusInternalServerError
	}

	updateRequestSlice := make([]map[string]interface{}, 0)
	if err = json.Unmarshal(updateRequestsBytes, &updateRequestSlice); err != nil {
		h.logger.Printf("[ERR] handler: Failed to unmarshal update requests to slice: %s", err.Error())
		status = http.StatusInternalServerError
	}

	updateRequests := make([]*protobuf.UpdateRequest, 0)
	for _, updateRequestMap := range updateRequestSlice {

		updateRequest := &protobuf.UpdateRequest{}
		if err = updateRequest.SetMap(updateRequestMap); err != nil {
			h.logger.Printf("[ERR] handler: Failed to get update request from map: %s", err.Error())
			status = http.StatusInternalServerError
			break
		}

		updateRequests = append(updateRequests, updateRequest)
	}

	req := &protobuf.BulkRequest{
		UpdateRequests: updateRequests,
		BatchSize:      int32(batchSize),
	}

	var resp *protobuf.BulkResponse
	if resp, err = h.client.Bulk(req); err != nil {
		h.logger.Printf("[ERR] handler: Failed to update in bulk: %s", err.Error())
		status = http.StatusInternalServerError
	}

	content := make([]byte, 0)
	if content, err = resp.GetBytes(); err != nil {
		h.logger.Printf("[ERR] handler: Failed to marshalling content: %s", err.Error())
		status = http.StatusInternalServerError
	}

	if prettyPrint {
		var buff bytes.Buffer
		if err = json.Indent(&buff, content, "", "  "); err != nil {
			h.logger.Printf("[ERR] handler: Failed to indent content: %s", err.Error())
			status = http.StatusInternalServerError
		}
		content = buff.Bytes()
	}

	// Write response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	if _, err = w.Write(content); err != nil {
		h.logger.Printf("[ERR] handler: Failed to write content: %s", err.Error())
	}

	return
}
