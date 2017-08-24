//  Copyright (c) 2017 Minoru Osuka
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
	"context"
	"encoding/json"
	"github.com/mosuka/blast/client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type GetIndexInfoHandler struct {
	client *client.Client
}

func NewGetIndexInfoHandler(client *client.Client) *GetIndexInfoHandler {
	return &GetIndexInfoHandler{
		client: client,
	}
}

func (h *GetIndexInfoHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"req": req,
	}).Info("")

	var indexPath bool
	if req.URL.Query().Get("indexMapping") == "true" {
		indexPath = true
	}

	var indexMapping bool
	if req.URL.Query().Get("indexMapping") == "true" {
		indexMapping = true
	}

	var indexType bool
	if req.URL.Query().Get("indexType") == "true" {
		indexType = true
	}

	var kvstore bool
	if req.URL.Query().Get("kvstore") == "true" {
		kvstore = true
	}

	var kvconfig bool
	if req.URL.Query().Get("kvconfig") == "true" {
		kvconfig = true
	}

	if !indexPath && !indexMapping && !indexType && !kvstore && !kvconfig {
		indexPath = true
		indexMapping = true
		indexType = true
		kvstore = true
		kvconfig = true
	}

	// request timeout
	requestTimeout := DefaultRequestTimeout
	if req.URL.Query().Get("requestTimeout") != "" {
		i, err := strconv.Atoi(req.URL.Query().Get("requestTimeout"))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to set batch size")

			Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requestTimeout = i
	}

	// create context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout)*time.Millisecond)
	defer cancel()

	// request
	resp, err := h.client.Index.GetIndexInfo(ctx, indexPath, indexMapping, indexType, kvstore, kvconfig)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get index")

		Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// output response
	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create response")

		Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(output)

	return
}
