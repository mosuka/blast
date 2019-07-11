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
	"reflect"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func TestDefaultIndexConfig(t *testing.T) {
	expConfig := &IndexConfig{
		IndexMapping:     mapping.NewIndexMapping(),
		IndexType:        bleve.Config.DefaultIndexType,
		IndexStorageType: bleve.Config.DefaultKVStore,
	}
	actConfig := DefaultIndexConfig()

	if !reflect.DeepEqual(expConfig, actConfig) {
		t.Fatalf("expected content to see %v, saw %v", expConfig, actConfig)
	}
}
