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

package protobuf

import (
	"encoding/json"
	"reflect"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/protobuf/raft"
	"github.com/mosuka/blast/registry"
)

func init() {
	registry.RegisterType("bool", reflect.TypeOf(false))
	registry.RegisterType("string", reflect.TypeOf(""))
	registry.RegisterType("int", reflect.TypeOf(int(0)))
	registry.RegisterType("int8", reflect.TypeOf(int8(0)))
	registry.RegisterType("int16", reflect.TypeOf(int16(0)))
	registry.RegisterType("int32", reflect.TypeOf(int32(0)))
	registry.RegisterType("int64", reflect.TypeOf(int64(0)))
	registry.RegisterType("uint", reflect.TypeOf(uint(0)))
	registry.RegisterType("uint8", reflect.TypeOf(uint8(0)))
	registry.RegisterType("uint16", reflect.TypeOf(uint16(0)))
	registry.RegisterType("uint32", reflect.TypeOf(uint32(0)))
	registry.RegisterType("uint64", reflect.TypeOf(uint64(0)))
	registry.RegisterType("uintptr", reflect.TypeOf(uintptr(0)))
	registry.RegisterType("byte", reflect.TypeOf(byte(0)))
	registry.RegisterType("rune", reflect.TypeOf(rune(0)))
	registry.RegisterType("float32", reflect.TypeOf(float32(0)))
	registry.RegisterType("float64", reflect.TypeOf(float64(0)))
	registry.RegisterType("complex64", reflect.TypeOf(complex64(0)))
	registry.RegisterType("complex128", reflect.TypeOf(complex128(0)))
	registry.RegisterType("map[string]interface {}", reflect.TypeOf((map[string]interface{})(nil)))
	registry.RegisterType("[]interface {}", reflect.TypeOf(([]interface{})(nil)))

	registry.RegisterType("index.Document", reflect.TypeOf(index.Document{}))
	registry.RegisterType("management.KeyValuePair", reflect.TypeOf(management.KeyValuePair{}))
	registry.RegisterType("raft.Node", reflect.TypeOf(raft.Node{}))

	registry.RegisterType("mapping.IndexMappingImpl", reflect.TypeOf(mapping.IndexMappingImpl{}))
	registry.RegisterType("bleve.SearchRequest", reflect.TypeOf(bleve.SearchRequest{}))
	registry.RegisterType("bleve.SearchResult", reflect.TypeOf(bleve.SearchResult{}))
}

func MarshalAny(message *any.Any) (interface{}, error) {
	if message == nil {
		return nil, nil
	}

	typeUrl := message.TypeUrl
	value := message.Value

	instance := registry.TypeInstanceByName(typeUrl)

	err := json.Unmarshal(value, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func UnmarshalAny(instance interface{}, message *any.Any) error {
	var err error

	if instance == nil {
		return nil
	}

	message.TypeUrl = registry.TypeNameByInstance(instance)

	message.Value, err = json.Marshal(instance)
	if err != nil {
		return err
	}

	return nil
}
