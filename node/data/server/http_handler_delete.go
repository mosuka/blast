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
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
)

type DeleteHandler struct {
	logger *log.Logger
	client *client.GRPCClient
}

func NewDeleteHandler(logger *log.Logger, client *client.GRPCClient) *DeleteHandler {
	return &DeleteHandler{
		logger: logger,
		client: client,
	}
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	start := time.Now()
	status := http.StatusOK
	defer HTTPMetrics(start, status, w, r, h.logger)

	vars := mux.Vars(r)
	id := vars["id"]

	prettyPrint, err := strconv.ParseBool(r.URL.Query().Get("pretty-print"))

	req := &protobuf.DeleteRequest{
		Id: id,
	}

	var resp *protobuf.DeleteResponse
	if resp, err = h.client.Delete(req); err != nil {
		h.logger.Printf("[ERR] handler: Failed to delete document: %s", err.Error())
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
