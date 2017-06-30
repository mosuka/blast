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
	"encoding/json"
	"github.com/mosuka/blast/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"net/http"
)

type GetIndexHandler struct {
	client *client.BlastClientWrapper
}

func NewGetIndexHandler(c *client.BlastClientWrapper) *GetIndexHandler {
	return &GetIndexHandler{
		client: c,
	}
}

func (h *GetIndexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"req": req,
	}).Info("")

	var includeIndexMapping bool
	if req.URL.Query().Get("includeIndexMapping") == "true" {
		includeIndexMapping = true
	}

	var includeIndexType bool
	if req.URL.Query().Get("includeIndexType") == "true" {
		includeIndexType = true
	}

	var includeKvstore bool
	if req.URL.Query().Get("includeKvstore") == "true" {
		includeKvstore = true
	}

	var includeKvconfig bool
	if req.URL.Query().Get("includeKvconfig") == "true" {
		includeKvconfig = true
	}

	// request
	resp, err := h.client.GetIndex(context.Background(), includeIndexMapping, includeIndexType, includeKvstore, includeKvconfig)
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
