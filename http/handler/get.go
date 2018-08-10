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
	"log"
	"net/http"
	"strconv"
	"time"

	"fmt"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/grpc/client"
	"github.com/mosuka/blast/http/metrics"
	"github.com/mosuka/blast/protobuf"
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
	var err error

	start := time.Now()
	status := http.StatusOK
	defer metrics.Metrics(start, status, w, r, h.logger)

	vars := mux.Vars(r)
	id := vars["id"]

	prettyPrint, err := strconv.ParseBool(r.URL.Query().Get("pretty-print"))

	req := &protobuf.GetRequest{
		Id: id,
	}

	var resp *protobuf.GetResponse
	if resp, err = h.client.Get(req); err != nil {
		status = http.StatusNotFound
	} else if err != nil {
		message := fmt.Sprintf("Failed to get fields: %s", err.Error())
		h.logger.Print(fmt.Sprintf("[ERR] %s", message))
		resp.Success = false
		resp.Message = message
		status = http.StatusInternalServerError
	}

	content := make([]byte, 0)
	if content, err = resp.GetBytes(); err != nil {
		message := fmt.Sprintf("Failed to marshalling get response: %s", err.Error())
		h.logger.Print(fmt.Sprintf("[ERR] %s", message))
		resp.Success = false
		resp.Message = message
		status = http.StatusInternalServerError
	}

	if prettyPrint {
		var buff bytes.Buffer
		if err = json.Indent(&buff, content, "", "  "); err != nil {
			message := fmt.Sprintf("Failed to indent content: %s", err.Error())
			h.logger.Print(fmt.Sprintf("[ERR] %s", message))
			resp.Success = false
			resp.Message = message
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
