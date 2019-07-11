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
	"reflect"
	"testing"

	"github.com/blevesearch/bleve/mapping"
)

func TestIndexConfig_SetIndexMapping(t *testing.T) {
	expIndexMapping := mapping.NewIndexMapping()
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexMapping(expIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexMapping, err := cfg.GetIndexMapping()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(expIndexMapping, actIndexMapping) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping, actIndexMapping)
	}
}

func TestIndexConfig_GetIndexMapping(t *testing.T) {
	expIndexMapping := mapping.NewIndexMapping()
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexMapping(expIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexMapping, err := cfg.GetIndexMapping()
	if err != nil {
		t.Fatalf("%v", err)
	}

	exp, err := json.Marshal(expIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}

	act, err := json.Marshal(actIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(exp, act) {
		t.Errorf("expected content to see %v, saw %v", exp, act)
	}
}

func TestIndexConfig_SetIndexType(t *testing.T) {
	expIndexType := "upside_down"
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexType(expIndexType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexType, err := cfg.GetIndexType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expIndexType != actIndexType {
		t.Fatalf("expected content to see %v, saw %v", expIndexType, actIndexType)
	}
}

func TestIndexConfig_GetIndexType(t *testing.T) {
	expIndexType := "upside_down"
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexType(expIndexType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexType, err := cfg.GetIndexType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expIndexType != actIndexType {
		t.Fatalf("expected content to see %v, saw %v", expIndexType, actIndexType)
	}
}

func TestIndexConfig_SetIndexStorageType(t *testing.T) {
	expIndexStorageType := "boltdb"
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexStorageType(expIndexStorageType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexStorageType, err := cfg.GetIndexStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expIndexStorageType != actIndexStorageType {
		t.Fatalf("expected content to see %v, saw %v", expIndexStorageType, actIndexStorageType)
	}
}

func TestIndexConfig_GetIndexStorageType(t *testing.T) {
	expIndexStorageType := "upside_down"
	cfg := DefaultIndexConfig()
	err := cfg.SetIndexStorageType(expIndexStorageType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexStorageType, err := cfg.GetIndexStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expIndexStorageType != actIndexStorageType {
		t.Fatalf("expected content to see %v, saw %v", expIndexStorageType, actIndexStorageType)
	}
}
