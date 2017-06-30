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

package util

import (
	"encoding/json"
	"github.com/blevesearch/bleve/mapping"
	"io"
	"io/ioutil"
)

func NewIndexMapping(reader io.Reader) (*mapping.IndexMappingImpl, error) {
	indexMapping := mapping.NewIndexMapping()

	resourceBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resourceBytes, indexMapping)
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func NewKvconfig(reader io.Reader) (map[string]interface{}, error) {
	kvconfig := make(map[string]interface{})

	resourceBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resourceBytes, &kvconfig)
	if err != nil {
		return nil, err
	}

	return kvconfig, nil
}
