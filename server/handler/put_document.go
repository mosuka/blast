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
	"github.com/buger/jsonparser"
	"github.com/gorilla/mux"
	"github.com/mosuka/blast/client"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type PutDocumentHandler struct {
	client *client.BlastClientWrapper
}

func NewPutDocumentHandler(c *client.BlastClientWrapper) *PutDocumentHandler {
	return &PutDocumentHandler{
		client: c,
	}
}

func (h *PutDocumentHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"req": req,
	}).Info("")

	vars := mux.Vars(req)

	// read request
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get id
	id, err := jsonparser.GetString(data, "document", "id")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get id")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get fields
	fieldsBytes, _, _, err := jsonparser.Get(data, "document", "fields")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get fields")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var fields map[string]interface{}
	err = json.Unmarshal(fieldsBytes, &fields)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create fields")

		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// overwrite request
	id = vars["id"]

	// request
	resp, err := h.client.PutDocument(id, fields)
	if err != nil {
		log.WithFields(log.Fields{
			"req": req,
		}).Error("failed to put document")

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
