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
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve/mapping"
)

func NewIndexMappingFromBytes(indexMappingBytes []byte) (*mapping.IndexMappingImpl, error) {
	indexMapping := mapping.NewIndexMapping()

	err := indexMapping.UnmarshalJSON(indexMappingBytes)
	if err != nil {
		return nil, err
	}

	err = indexMapping.Validate()
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func NewIndexMappingFromMap(indexMappingMap map[string]interface{}) (*mapping.IndexMappingImpl, error) {
	indexMappingBytes, err := json.Marshal(indexMappingMap)
	if err != nil {
		return nil, err
	}

	indexMapping, err := NewIndexMappingFromBytes(indexMappingBytes)
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func NewIndexMappingFromFile(indexMappingPath string) (*mapping.IndexMappingImpl, error) {
	_, err := os.Stat(indexMappingPath)
	if err != nil {
		if os.IsNotExist(err) {
			// does not exist
			return nil, err
		}
		// other error
		return nil, err
	}

	// read index mapping file
	indexMappingFile, err := os.Open(indexMappingPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = indexMappingFile.Close()
	}()

	indexMappingBytes, err := ioutil.ReadAll(indexMappingFile)
	if err != nil {
		return nil, err
	}

	indexMapping, err := NewIndexMappingFromBytes(indexMappingBytes)
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}
