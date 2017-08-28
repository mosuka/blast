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
	"github.com/blevesearch/bleve"
	"github.com/buger/jsonparser"
	"github.com/mosuka/blast/client"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SearchHandler struct {
	client *client.BlastClient
}

func NewSearchHandler(c *client.BlastClient) *SearchHandler {
	return &SearchHandler{
		client: c,
	}
}

type SearchResponse struct {
	SearchResult map[string]interface{} `json:"search_result"`
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"req": req,
	}).Info("")

	// read request
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get search_request
	searchRequestBytes, _, _, err := jsonparser.Get(data, "search_request")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read search request")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var searchRequest *bleve.SearchRequest
	err = json.Unmarshal(searchRequestBytes, &searchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create search request")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// overwrite request
	if req.URL.Query().Get("query") != "" {
		searchRequest.Query = bleve.NewQueryStringQuery(req.URL.Query().Get("query"))
	}
	if req.URL.Query().Get("size") != "" {
		i, err := strconv.Atoi(req.URL.Query().Get("size"))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to convert size")

			Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		searchRequest.Size = i
	}
	if searchRequest.Size == 0 {
		searchRequest.Size = DefaultSize
	}
	if req.URL.Query().Get("from") != "" {
		i, err := strconv.Atoi(req.URL.Query().Get("from"))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to convert from")

			Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		searchRequest.From = i
	}
	if searchRequest.From < 0 {
		searchRequest.From = DefaultFrom
	}
	if req.URL.Query().Get("explain") != "" {
		if req.URL.Query().Get("explain") == "true" {
			searchRequest.Explain = true
		} else {
			searchRequest.Explain = false
		}
	}
	if req.URL.Query().Get("fields") != "" {
		searchRequest.Fields = strings.Split(req.URL.Query().Get("fields"), ",")
	}
	if req.URL.Query().Get("sort") != "" {
		searchRequest.SortBy(strings.Split(req.URL.Query().Get("sort"), ","))
	}
	if req.URL.Query().Get("facets") != "" {
		facetRequest := bleve.FacetsRequest{}
		err := json.Unmarshal([]byte(req.URL.Query().Get("facets")), &facetRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to create facets")

			Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		searchRequest.Facets = facetRequest
	}
	if req.URL.Query().Get("highlight") != "" {
		highlightRequest := bleve.NewHighlight()
		err := json.Unmarshal([]byte(req.URL.Query().Get("highlight")), highlightRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to create highlight")

			Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		searchRequest.Highlight = highlightRequest
	}
	if req.URL.Query().Get("highlightStyle") != "" || req.URL.Query().Get("highlightField") != "" {
		highlightRequest := bleve.NewHighlightWithStyle(req.URL.Query().Get("highlightStyle"))
		highlightRequest.Fields = strings.Split(req.URL.Query().Get("highlightField"), ",")
		searchRequest.Highlight = highlightRequest
	}
	if req.URL.Query().Get("include-locations") != "" {
		if req.URL.Query().Get("include-locations") == "true" {
			searchRequest.IncludeLocations = true
		} else {
			searchRequest.IncludeLocations = false
		}
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
	resp, err := h.client.Index.Search(ctx, searchRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"req": req,
		}).Error("failed to search documents")

		Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// output response
	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.WithFields(log.Fields{
			"req": req,
		}).Error("failed to create response")

		Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(output)

	return
}
