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

package maputils

import (
	"reflect"
	"testing"

	"github.com/stretchr/objx"
)

func TestStructuredMap_LoadJSON(t *testing.T) {
	m := NewStructuredMap()

	data := []byte(`
{
 "a": "A",
 "b": {
   "c": "B-C",
   "d": "B-D"
 }
}
`)

	err := m.LoadJSON(data)
	if err != nil {
		t.Errorf("%v", m)
	}
	val1 := m.data
	exp1 := objx.Map{
		"a": "A",
		"b": objx.Map{
			"c": "B-C",
			"d": "B-D",
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	data = []byte(`
[
  "a",
  1
]
`)

	err = m.LoadJSON(data)
	if err == nil {
		t.Error("expected nil")
	}
}

func TestStructured_LoadYAML(t *testing.T) {
	m := NewStructuredMap()

	data := []byte(`
---
a: A
b:
  c: B-C
  d: B-D
`)

	err := m.LoadYAML(data)
	if err != nil {
		t.Errorf("%v", m)
	}
	val1 := m.data
	exp1 := objx.Map{
		"a": "A",
		"b": objx.Map{
			"c": "B-C",
			"d": "B-D",
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	data = []byte(`
[
  "a",
  1
]
`)

	err = m.LoadYAML(data)
	if err == nil {
		t.Error("expected nil")
	}
}

func TestStructuredMap_Get(t *testing.T) {
	m := NewStructuredMap()

	data := []byte(`
{
  "a": "A",
  "b": {
    "c": "B-C",
    "d": "B-D"
  },
  "c": [
    "C-0",
    "C-1"
  ],
  "d": [
    {"e": "D-0-E", "g": "D-0-G"},
    {"f": "D-1-F", "h": "D-1-H"}
  ]
}
`)

	err := m.LoadJSON(data)
	if err != nil {
		t.Errorf("%v", m)
	}

	val1, err := m.Get("/a")
	if err != nil {
		t.Errorf("%v", m)
	}
	exp1 := "A"
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	val2, err := m.Get("/b")
	if err != nil {
		t.Errorf("%v", m)
	}
	exp2 := objx.Map{
		"c": "B-C",
		"d": "B-D",
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	val3, err := m.Get("/c[1]")
	if err != nil {
		t.Errorf("%v", m)
	}
	exp3 := "C-1"
	act3 := val3
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", act3, act3)
	}

	val4, err := m.Get("/d[1]/h")
	if err != nil {
		t.Errorf("%v", m)
	}
	exp4 := "D-1-H"
	act4 := val4
	if !reflect.DeepEqual(exp4, act4) {
		t.Errorf("expected content to see %v, saw %v", act4, act4)
	}
}

func TestStructuredMap_Set(t *testing.T) {
	m := NewStructuredMap()

	err := m.Set("/", map[string]interface{}{"a": "A", "b": "B"})
	if err != nil {
		t.Errorf("%v", m)
	}

	err = m.Set("/a", "AAA")
	if err != nil {
		t.Errorf("%v", m)
	}
	exp1 := objx.Map{
		"a": "AAA",
		"b": "B",
	}
	act1 := m.data
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	err = m.Set("/b", []string{})
	if err != nil {
		t.Errorf("%v", m)
	}
	exp2 := objx.Map{
		"a": "AAA",
		"b": []string{},
	}
	act2 := m.data
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	err = m.Set("/b", []string{"B-0"})
	if err != nil {
		t.Errorf("%v", m)
	}
	exp3 := objx.Map{
		"a": "AAA",
		"b": []string{"B-0"},
	}
	act3 := m.data
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}
}

func TestStructuredMap_Delete(t *testing.T) {
	m := NewStructuredMap()

	data := []byte(`
{
 "a": "A",
 "b": {
   "c": "B-C",
   "d": "B-D"
 }
}
`)

	err := m.LoadJSON(data)
	if err != nil {
		t.Errorf("%v", err)
	}

	err = m.Delete("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := objx.Map{}
	act1 := m.data.Value().ObjxMap()
	if err != nil {
		t.Errorf("%v", err)
	}
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	err = m.LoadJSON(data)
	if err != nil {
		t.Errorf("%v", err)
	}

	err = m.Delete("/a")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := objx.Map{
		"b": objx.Map{
			"c": "B-C",
			"d": "B-D",
		},
	}
	act2 := m.data.Value().ObjxMap()
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	err = m.Delete("/b/d")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp3 := objx.Map{
		"b": objx.Map{
			"c": "B-C",
		},
	}
	act3 := m.data.Value().ObjxMap()
	if err != nil {
		t.Errorf("%v", err)
	}
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}

}
