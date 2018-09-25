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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
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
	start := time.Now()
	status := http.StatusOK
	defer HTTPMetrics(start, status, w, r, h.logger)

	var err error

	searchRequestBytes, err := ioutil.ReadAll(r.Body)
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

	if len(searchRequestBytes) <= 0 {
		err := errors.New("search request must be set")
		status = http.StatusBadRequest
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	searchRequest := bleve.NewSearchRequest(nil)
	if searchRequestBytes != nil {
		err = json.Unmarshal(searchRequestBytes, searchRequest)
		if err != nil {
			status = http.StatusBadRequest
			errContent, err := NewContent(err.Error())
			if err != nil {
				h.logger.Printf("[ERR] %v", err)
			}
			WriteResponse(w, errContent, status, h.logger)
			return
		}
	}

	searchRequestAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	if err != nil {
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	req := &protobuf.SearchDocumentsRequest{
		SearchRequest: searchRequestAny,
	}

	resp, err := h.client.SearchDocuments(req)
	if err != nil {
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	searchResultInstance, err := protobuf.MarshalAny(resp.SearchResult)
	if err != nil {
		status = http.StatusInternalServerError
		errContent, err := NewContent(err.Error())
		if err != nil {
			h.logger.Printf("[ERR] %v", err)
		}
		WriteResponse(w, errContent, status, h.logger)
		return
	}

	content, err := json.MarshalIndent(searchResultInstance, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	WriteResponse(w, content, status, h.logger)

	return
}
