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

package config

import (
	"encoding/json"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type IndexConfig struct {
	IndexMapping     *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
	IndexType        string                    `json:"index_type,omitempty"`
	IndexStorageType string                    `json:"index_storage_type,omitempty"`
}

func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		IndexMapping:     mapping.NewIndexMapping(),
		IndexType:        bleve.Config.DefaultIndexType,
		IndexStorageType: bleve.Config.DefaultKVStore,
	}
}

func NewIndexConfigFromMap(src map[string]interface{}) *IndexConfig {
	b, err := json.Marshal(src)
	if err != nil {
		return &IndexConfig{}
	}

	var indexConfig *IndexConfig
	err = json.Unmarshal(b, &indexConfig)
	if err != nil {
		return &IndexConfig{}
	}

	return indexConfig
}

func (c *IndexConfig) ToMap() map[string]interface{} {
	b, err := json.Marshal(c)
	if err != nil {
		return map[string]interface{}{}
	}

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return map[string]interface{}{}
	}

	return m
}
