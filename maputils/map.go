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
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/stretchr/objx"
)

type StructuredMap struct {
	data objx.Map
}

func NewStructuredMap() *StructuredMap {
	m := &StructuredMap{
		data: objx.Map{},
	}

	return m
}

func (m *StructuredMap) LoadJSON(data []byte) error {
	var err error
	m.data, err = objx.FromJSON(string(data))
	if err != nil {
		return err
	}

	return nil
}

func (m *StructuredMap) LoadYAML(data []byte) error {
	d, err := yaml.YAMLToJSON(data)

	err = m.LoadJSON(d)
	if err != nil {
		return err
	}

	return nil
}

func (m *StructuredMap) pathKeys(path string) []string {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	return keys
}

func (m *StructuredMap) makeSafePath(path string) string {
	keys := m.pathKeys(path)

	safePath := strings.Join(keys, "/")
	//safePath = strings.Join(strings.Split(safePath, "."), "_")
	safePath = strings.Join(strings.Split(safePath, "/"), objx.PathSeparator)

	return safePath
}

func (m *StructuredMap) Set(path string, value interface{}) error {
	if path == "/" {
		// convert to JSON string
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}

		err = m.LoadJSON(b)
		if err != nil {
			return err
		}
	} else {
		m.data.Set(m.makeSafePath(path), value)
	}

	return nil
}

func (m *StructuredMap) Get(path string) (interface{}, error) {
	var value *objx.Value

	if path == "/" {
		value = m.data.Value()
	} else {
		value = m.data.Get(m.makeSafePath(path))
	}

	if value.IsBool() {
		return value.Bool(), nil
	} else if value.IsBoolSlice() {
		return value.BoolSlice(), nil
	} else if value.IsComplex64() {
		return value.Complex64(), nil
	} else if value.IsComplex64Slice() {
		return value.Complex64Slice(), nil
	} else if value.IsComplex128() {
		return value.Complex128(), nil
	} else if value.IsComplex128Slice() {
		return value.Complex128Slice(), nil
	} else if value.IsFloat32() {
		return value.Float32(), nil
	} else if value.IsFloat32Slice() {
		return value.Float32Slice(), nil
	} else if value.IsFloat64() {
		return value.Float64(), nil
	} else if value.IsFloat64Slice() {
		return value.Float64Slice(), nil
	} else if value.IsInt() {
		return value.Int(), nil
	} else if value.IsIntSlice() {
		return value.IntSlice(), nil
	} else if value.IsInt8() {
		return value.Int8(), nil
	} else if value.IsInt8Slice() {
		return value.Int8Slice(), nil
	} else if value.IsInt16() {
		return value.Int16(), nil
	} else if value.IsInt16Slice() {
		return value.Int16Slice(), nil
	} else if value.IsInt32() {
		return value.Int32(), nil
	} else if value.IsInt32Slice() {
		return value.Int32Slice(), nil
	} else if value.IsInt64() {
		return value.Int64(), nil
	} else if value.IsInt64Slice() {
		return value.Int64Slice(), nil
	} else if value.IsStr() {
		return value.Str(), nil
	} else if value.IsStrSlice() {
		return value.StrSlice(), nil
	} else if value.IsUint() {
		return value.Uint(), nil
	} else if value.IsUintSlice() {
		return value.UintSlice(), nil
	} else if value.IsUint8() {
		return value.Uint8(), nil
	} else if value.IsUint8Slice() {
		return value.Uint8Slice(), nil
	} else if value.IsUint16() {
		return value.Uint16(), nil
	} else if value.IsUint16Slice() {
		return value.Uint16Slice(), nil
	} else if value.IsUint32() {
		return value.Uint32(), nil
	} else if value.IsUint32Slice() {
		return value.Uint32Slice(), nil
	} else if value.IsUint64() {
		return value.Uint64(), nil
	} else if value.IsUint64Slice() {
		return value.Uint64Slice(), nil
	} else if value.IsUintptr() {
		return value.Uintptr(), nil
	} else if value.IsUintptrSlice() {
		return value.UintptrSlice(), nil
	} else if value.IsObjxMap() {
		return value.ObjxMap(), nil
	} else if value.IsObjxMapSlice() {
		return value.ObjxMapSlice(), nil
	} else if value.IsMSI() {
		return value.MSI(), nil
	} else if value.IsMSISlice() {
		return value.MSISlice(), nil
	} else if value.IsInter() {
		return value.Inter(), nil
	} else if value.IsInterSlice() {
		return value.InterSlice(), nil
	} else if value.IsNil() {
		return nil, nil
	} else {
		return value, nil
	}
}

func (m *StructuredMap) delete(keys []string, data interface{}) (interface{}, error) {
	var err error

	if len(keys) >= 1 {
		key := keys[0]

		switch data.(type) {
		case map[string]interface{}:
			if _, exist := data.(map[string]interface{})[key]; exist {
				// key exists
				if len(keys) > 1 {
					data.(map[string]interface{})[key], err = m.delete(keys[1:], data.(map[string]interface{})[key])
					if err != nil {
						return nil, err
					}
				} else {
					//
					mm := data.(map[string]interface{})
					delete(mm, key)
					data = mm
				}
			} else {
				// key does not exist
				return nil, blasterrors.ErrNotFound
			}
		}
	}

	return data, nil
}

func (m *StructuredMap) Delete(path string) error {
	if path == "/" {
		m.data = objx.Map{}
		return nil
	}

	b, err := json.Marshal(&m.data)
	if err != nil {
		return err
	}
	var mm map[string]interface{}
	err = json.Unmarshal(b, &mm)
	if err != nil {
		return err
	}

	data, err := m.delete(m.pathKeys(path), mm)
	if err != nil {
		return err
	}

	b, err = json.Marshal(&data)
	if err != nil {
		return err
	}

	m.data, err = objx.FromJSON(string(b))
	if err != nil {
		return err
	}

	return nil
}
