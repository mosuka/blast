// Copyright (c) 2019 Minoru Osuka
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

package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func NewJSONMessage(msgMap map[string]interface{}) ([]byte, error) {
	content, err := json.MarshalIndent(msgMap, "", "  ")
	if err != nil {
		return nil, err
	}

	return content, nil
}

func WriteResponse(w http.ResponseWriter, content []byte, status int, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	_, err := w.Write(content)
	if err != nil {
		logger.Printf("[ERR] handler: Failed to write content: %s", err.Error())
	}

	return
}
