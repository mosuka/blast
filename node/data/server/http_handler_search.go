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
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
)

type SearchHandler struct {
	logger *log.Logger
	client *client.GRPCClient
}

func NewSearchHandler(logger *log.Logger, client *client.GRPCClient) *SearchHandler {
	return &SearchHandler{
		logger: logger,
		client: client,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	start := time.Now()
	status := http.StatusOK
	defer HTTPMetrics(start, status, w, r, h.logger)

	prettyPrint, err := strconv.ParseBool(r.URL.Query().Get("pretty-print"))

	searchRequestBytes := make([]byte, 0)
	if searchRequestBytes, err = ioutil.ReadAll(r.Body); err != nil {
		h.logger.Printf("[ERR] handler: Failed to read request body: %s", err.Error())
		status = http.StatusInternalServerError
	}

	req := &protobuf.SearchRequest{}
	if err = req.SetBytes(searchRequestBytes); err != nil {
		h.logger.Printf("[ERR] handler: Failed to create search request: %s", err.Error())
		status = http.StatusInternalServerError
	}

	var resp *protobuf.SearchResponse
	if resp, err = h.client.Search(req); err != nil {
		status = http.StatusNotFound
	} else if err != nil {
		h.logger.Printf("[ERR] handler: Failed to search documents: %s", err.Error())
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
