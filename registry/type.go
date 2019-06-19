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

package registry

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func init() {
	RegisterType("bool", reflect.TypeOf(false))
	RegisterType("string", reflect.TypeOf(""))
	RegisterType("int", reflect.TypeOf(int(0)))
	RegisterType("int8", reflect.TypeOf(int8(0)))
	RegisterType("int16", reflect.TypeOf(int16(0)))
	RegisterType("int32", reflect.TypeOf(int32(0)))
	RegisterType("int64", reflect.TypeOf(int64(0)))
	RegisterType("uint", reflect.TypeOf(uint(0)))
	RegisterType("uint8", reflect.TypeOf(uint8(0)))
	RegisterType("uint16", reflect.TypeOf(uint16(0)))
	RegisterType("uint32", reflect.TypeOf(uint32(0)))
	RegisterType("uint64", reflect.TypeOf(uint64(0)))
	RegisterType("uintptr", reflect.TypeOf(uintptr(0)))
	RegisterType("byte", reflect.TypeOf(byte(0)))
	RegisterType("rune", reflect.TypeOf(rune(0)))
	RegisterType("float32", reflect.TypeOf(float32(0)))
	RegisterType("float64", reflect.TypeOf(float64(0)))
	RegisterType("complex64", reflect.TypeOf(complex64(0)))
	RegisterType("complex128", reflect.TypeOf(complex128(0)))

	RegisterType("map[string]interface {}", reflect.TypeOf((map[string]interface{})(nil)))
	RegisterType("[]interface {}", reflect.TypeOf(([]interface{})(nil)))

	RegisterType("mapping.IndexMappingImpl", reflect.TypeOf(mapping.IndexMappingImpl{}))
	RegisterType("bleve.SearchRequest", reflect.TypeOf(bleve.SearchRequest{}))
	RegisterType("bleve.SearchResult", reflect.TypeOf(bleve.SearchResult{}))
}

type TypeRegistry map[string]reflect.Type

var Types = make(TypeRegistry, 0)

func RegisterType(name string, typ reflect.Type) {
	if _, exists := Types[name]; exists {
		panic(errors.New(fmt.Sprintf("attempted to register duplicate index: %s", name)))
	}
	Types[name] = typ
}

func TypeByName(name string) reflect.Type {
	return Types[name]
}

func TypeNameByInstance(instance interface{}) string {
	switch instance.(type) {
	case bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		return reflect.TypeOf(instance).Name()
	case map[string]interface{}, []interface{}:
		return reflect.TypeOf(instance).String()
	default:
		return reflect.TypeOf(instance).Elem().String()
	}
}

func TypeInstanceByName(name string) interface{} {
	return reflect.New(TypeByName(name)).Interface()
}
