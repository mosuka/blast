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

	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/maputils"
)

type IndexConfig struct {
	maputils.Map
}

func DefaultIndexConfig() IndexConfig {
	c := IndexConfig{Map: maputils.New()}

	_ = c.SetIndexMapping(mapping.NewIndexMapping())
	_ = c.SetIndexType("upside_down")
	_ = c.SetIndexStorageType("boltdb")

	return c
}

func NewIndexConfigFromMap(src map[string]interface{}) IndexConfig {
	return IndexConfig{Map: maputils.FromMap(src)}
}

func (c IndexConfig) SetIndexMapping(indexMapping *mapping.IndexMappingImpl) error {
	err := indexMapping.Validate()
	if err != nil {
		return err
	}

	// IndexMappingImpl -> JSON
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		return err
	}

	// JSON -> map[string]interface{}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		return err
	}

	err = c.Set("/index_mapping", indexMappingMap)
	if err != nil {
		return err
	}
	return nil
}

func (c IndexConfig) GetIndexMapping() (*mapping.IndexMappingImpl, error) {
	indexMappingMap, err := c.Get("/index_mapping")
	if err != nil {
		return nil, err
	}

	indexMapping, err := indexutils.NewIndexMappingFromMap(indexMappingMap.(maputils.Map).ToMap())
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func (c IndexConfig) SetIndexType(indexType string) error {
	err := c.Set("/index_type", indexType)
	if err != nil {
		return err
	}
	return nil
}

func (c IndexConfig) GetIndexType() (string, error) {
	indexType, err := c.Get("/index_type")
	if err != nil {
		return "", err
	}
	return indexType.(string), nil
}

func (c IndexConfig) SetIndexStorageType(indexStorageType string) error {
	err := c.Set("/index_storage_type", indexStorageType)
	if err != nil {
		return err
	}
	return nil
}

func (c IndexConfig) GetIndexStorageType() (string, error) {
	indexStorageType, err := c.Get("/index_storage_type")
	if err != nil {
		return "", err
	}
	return indexStorageType.(string), nil
}
