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
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mosuka/blast/version"
)

type IndexHandler struct {
	logger *log.Logger
}

func NewIndexHandler(logger *log.Logger) *IndexHandler {
	return &IndexHandler{
		logger: logger,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	start := time.Now()
	status := http.StatusOK
	defer HTTPMetrics(start, status, w, r, h.logger)

	// Create content
	content := []byte(`<html>
                      <head><title>Blast ` + version.Version + `</title></head>
                      <body>
                      <h1>Blast ` + version.Version + `</h1>
                      </body>
                      </html>`)

	// Write response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	if _, err = w.Write(content); err != nil {
		h.logger.Printf("[ERR] handler: Failed to write content: %s", err.Error())
	}

	return
}
