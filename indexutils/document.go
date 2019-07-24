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

package indexutils

import (
	"encoding/json"
	"errors"
)

type Document struct {
	Id     string                 `json:"id,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

func NewDocument(id string, fields map[string]interface{}) (*Document, error) {
	doc := &Document{
		Id:     id,
		Fields: fields,
	}

	if err := doc.Validate(); err != nil {
		return nil, err
	}

	return doc, nil
}

func NewDocumentFromBytes(src []byte) (*Document, error) {
	var doc *Document

	err := json.Unmarshal(src, &doc)
	if err != nil {
		return nil, err
	}

	if err := doc.Validate(); err != nil {
		return nil, err
	}

	return doc, nil
}

func (d *Document) Validate() error {
	if d.Id == "" {
		return errors.New("id is empty")
	}

	if len(d.Fields) <= 0 {
		return errors.New("fields are empty")
	}

	return nil
}
