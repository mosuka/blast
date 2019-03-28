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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
)

type NestedMap struct {
	data interface{}
	re   *regexp.Regexp
}

func NewNestedMap(data interface{}) (*NestedMap, error) {
	switch data.(type) {
	case map[string]interface{}:
	default:
		return nil, errors.New(fmt.Sprintf("not map"))
	}

	return &NestedMap{
		data: data,
		re:   regexp.MustCompile("^(.+)\\[(\\d*)\\]$"),
	}, nil
}

func (m *NestedMap) Get(path string) (interface{}, error) {
	var err error

	value := m.data

	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	for _, key := range keys {
		if key == "" {
			continue
		}

		i := -1
		group := m.re.FindStringSubmatch(key)
		if len(group) >= 3 {
			key = group[1]
			i, err = strconv.Atoi(group[2])
			if err != nil {
				return nil, err
			}
			if i < 0 {
				return nil, errors.New(fmt.Sprintf("index out of bounds: %d", i))
			}
		}

		switch value.(type) {
		case map[string]interface{}:
			var exists bool
			value, exists = value.(map[string]interface{})[key]
			if !exists {
				return nil, ErrNotFound
			}
			if i >= 0 {
				arr := value.([]interface{})
				if i >= len(arr) {
					return nil, errors.New(fmt.Sprintf("index out of bounds: [index:%d] [array:%v]", i, arr))
				}
				value = arr[i]
			}
		}
	}

	return value, nil
}

func (m *NestedMap) makeMap(path string, value interface{}) map[string]interface{} {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	if len(keys) > 1 {
		value = m.makeMap(strings.Join(keys[1:], "/"), value)
	}

	return map[string]interface{}{keys[0]: value}
}

func (m *NestedMap) Set(path string, value interface{}) error {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	dst := m.data

	iter := newIterator(keys)
	for {
		key, err := iter.value()
		if err != nil {
			return err
		}
		if _, exists := dst.(map[string]interface{})[key]; !exists {
			if !iter.hasNext() {
				dst.(map[string]interface{})[key] = value
				return nil
			}

			dst.(map[string]interface{})[key] = map[string]interface{}{}
			dst = dst.(map[string]interface{})[key]
		} else {
			switch dst.(map[string]interface{})[key].(type) {
			case map[string]interface{}:

				switch value.(type) {
				case map[string]interface{}:
					err := mergo.Merge(&dst, value, mergo.WithOverride)
					if err != nil {
						return err
					}
				case []interface{}, interface{}:
					dst.(map[string]interface{})[key] = value
				}

			case []interface{}, interface{}:
				dst.(map[string]interface{})[key] = value
			}
		}

		if iter.hasNext() {
			iter.next()
		} else {
			break
		}
	}

	return nil
}

func (m *NestedMap) Delete(path string) error {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	if len(keys) > 1 {
		err := m.Delete(strings.Join(keys[1:], "/"))
		if err != nil {
			return err
		}
	}

	delete(m.data.(map[string]interface{}), keys[0])

	return nil
}

type iterator struct {
	keys []string
	pos  int
}

func newIterator(keys []string) *iterator {
	return &iterator{
		keys: keys,
		pos:  0,
	}
}

func (i *iterator) hasNext() bool {
	return i.pos < len(i.keys)-1
}

func (i *iterator) next() bool {
	i.pos++
	return i.pos < len(i.keys)-1
}

func (i *iterator) value() (string, error) {
	if i.pos > len(i.keys)-1 {
		return "", errors.New("value is not valid after iterator finished")
	}
	return i.keys[i.pos], nil
}
