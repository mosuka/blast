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

package index

import (
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/protobuf"
)

func MarshalDocument(doc *Document) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("nil")
	}

	fieldsIntr, err := protobuf.MarshalAny(doc.Fields)
	if err != nil {
		return nil, err
	}

	docMap := map[string]interface{}{
		"id":     doc.Id,
		"fields": *fieldsIntr.(*map[string]interface{}),
	}

	docBytes, err := json.Marshal(docMap)
	if err != nil {
		return nil, err
	}

	return docBytes, nil
}

func UnmarshalDocument(data []byte, doc *Document) error {
	var err error

	if data == nil || len(data) <= 0 || doc == nil {
		return nil
	}

	var docMap map[string]interface{}
	err = json.Unmarshal(data, &docMap)
	if err != nil {
		return err
	}

	if id, ok := docMap["id"].(string); ok {
		doc.Id = id
	}

	if fieldsMap, ok := docMap["fields"].(map[string]interface{}); ok {
		fieldsAny := &any.Any{}
		err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
		if err != nil {
			return err
		}
		doc.Fields = fieldsAny
	}

	return nil
}
