package maputils

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/imdario/mergo"
	"github.com/stretchr/objx"
	yaml "gopkg.in/yaml.v2"
)

type Map map[string]interface{}

func FromMap(src map[string]interface{}) Map {
	return normalize(src).(Map)
}

func FromJSON(src []byte) (Map, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal(src, &m)
	if err != nil {
		return nil, err
	}

	return FromMap(m), nil
}

func FromYAML(src []byte) (Map, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(src, &m)
	if err != nil {
		return nil, err
	}

	return FromMap(m), nil
}

func splitKey(path string) []string {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	return keys
}

func makeSelector(key string) string {
	return strings.Join(splitKey(key), objx.PathSeparator)
}

func normalize(value interface{}) interface{} {
	switch value.(type) {
	case map[string]interface{}:
		ret := Map{}
		for k, v := range value.(map[string]interface{}) {
			ret[k] = normalize(v)
		}
		return ret
	case map[interface{}]interface{}: // when unmarshaled by yaml
		ret := Map{}
		for k, v := range value.(map[interface{}]interface{}) {
			ret[k.(string)] = normalize(v)
		}
		return ret
	case []interface{}:
		ret := make([]interface{}, 0)
		for _, v := range value.([]interface{}) {
			ret = append(ret, normalize(v))
		}
		return ret
	case bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		return value
	default:
		return value
	}
}

func _makeMap(path string, value interface{}) interface{} {
	var ret interface{}

	keys := splitKey(path)

	if len(keys) >= 1 {
		ret = Map{keys[0]: _makeMap(strings.Join(keys[1:], "/"), value)}
	} else if len(keys) == 0 {
		ret = normalize(value)
	}

	return ret
}

func makeMap(path string, value interface{}) Map {
	mm := _makeMap(path, value)

	if _, ok := mm.(Map); !ok {
		return nil
	}

	return mm.(Map)
}

func (m Map) Has(key string) (bool, error) {
	value, err := m.Get(key)
	if err == ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return value != nil, nil
}

func (m Map) Set(key string, value interface{}) error {
	exist, err := m.Has(key)
	if err != ErrNotFound && err != nil {
		return err
	}

	if exist {
		err = m.Delete(key)
		if err != nil {
			return err
		}
	}

	mm := makeMap(key, value)

	err = mergo.Merge(&m, mm, mergo.WithOverride)
	if err != nil {
		return err
	}

	return nil
}

func (m Map) Merge(key string, value interface{}) error {
	mm := makeMap(key, value)

	err := mergo.Merge(&m, mm)
	if err != nil {
		return err
	}

	return nil
}

func (m Map) Get(key string) (interface{}, error) {
	var tmpMap interface{}

	var value interface{}
	tmpMap = m
	iter := newIterator(splitKey(key))
	for {
		k, err := iter.value()
		if err != nil {
			return nil, err
		}

		if _, ok := tmpMap.(Map)[k]; !ok {
			return nil, ErrNotFound
		}

		if iter.hasNext() {
			tmpMap = tmpMap.(Map)[k]
			iter.next()
		} else {
			value = tmpMap.(Map)[k]
			break
		}
	}

	return value, nil
}

func (m Map) Delete(key string) error {
	var tmpMap interface{}

	tmpMap = m
	iter := newIterator(splitKey(key))
	for {
		k, err := iter.value()
		if err != nil {
			return err
		}

		if _, ok := tmpMap.(Map)[k]; !ok {
			return ErrNotFound
		}

		if iter.hasNext() {
			tmpMap = tmpMap.(Map)[k]
			iter.next()
		} else {
			delete(tmpMap.(Map), k)
			break
		}
	}

	return nil
}

func (m Map) toMap(value interface{}) interface{} {
	switch value.(type) {
	case Map:
		ret := map[string]interface{}{}
		for k, v := range value.(Map) {
			ret[k] = m.toMap(v)
		}
		return ret
	case []interface{}:
		ret := make([]interface{}, 0)
		for _, v := range value.([]interface{}) {
			ret = append(ret, m.toMap(v))
		}
		return ret
	case bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		return value
	default:
		return value
	}
}

func (m Map) ToMap() map[string]interface{} {
	return m.toMap(m).(map[string]interface{})
}

func (m Map) ToJSON() ([]byte, error) {
	mm := m.ToMap()
	b, err := json.Marshal(&mm)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (m Map) ToYAML() ([]byte, error) {
	mm := m.ToMap()
	b, err := yaml.Marshal(&mm)
	if err != nil {
		return nil, err
	}

	return b, nil
}

//func Get(src map[string]interface{}, key string) (interface{}, error) {
//	tmpMap := objx.New(src)
//	var tmpValue *objx.Value
//
//	selector := makeSelector(key)
//	if selector == "" {
//		tmpValue = tmpMap.Value()
//	} else {
//		if tmpMap.Has(selector) {
//			tmpValue = tmpMap.Get(selector)
//		} else {
//			return nil, ErrNotFound
//		}
//	}
//
//	var value interface{}
//	if tmpValue.IsBool() {
//		value = tmpValue.Bool()
//	} else if tmpValue.IsBoolSlice() {
//		value = tmpValue.BoolSlice()
//	} else if tmpValue.IsComplex64() {
//		value = tmpValue.Complex64()
//	} else if tmpValue.IsComplex64Slice() {
//		value = tmpValue.Complex64Slice()
//	} else if tmpValue.IsComplex128() {
//		value = tmpValue.Complex128()
//	} else if tmpValue.IsComplex128Slice() {
//		value = tmpValue.Complex128Slice()
//	} else if tmpValue.IsFloat32() {
//		value = tmpValue.Float32()
//	} else if tmpValue.IsFloat32Slice() {
//		value = tmpValue.Float32Slice()
//	} else if tmpValue.IsFloat64() {
//		value = tmpValue.Float64()
//	} else if tmpValue.IsFloat64Slice() {
//		value = tmpValue.Float64Slice()
//	} else if tmpValue.IsInt() {
//		value = tmpValue.Int()
//	} else if tmpValue.IsIntSlice() {
//		value = tmpValue.IntSlice()
//	} else if tmpValue.IsInt8() {
//		value = tmpValue.IsInt8()
//	} else if tmpValue.IsInt8Slice() {
//		value = tmpValue.Int8Slice()
//	} else if tmpValue.IsInt16() {
//		value = tmpValue.Int16()
//	} else if tmpValue.IsInt16Slice() {
//		value = tmpValue.Int16Slice()
//	} else if tmpValue.IsInt32() {
//		value = tmpValue.Int32()
//	} else if tmpValue.IsInt32Slice() {
//		value = tmpValue.Int32Slice()
//	} else if tmpValue.IsInt64() {
//		value = tmpValue.IsInt64()
//	} else if tmpValue.IsInt64Slice() {
//		value = tmpValue.Int64Slice()
//	} else if tmpValue.IsStr() {
//		value = tmpValue.Str()
//	} else if tmpValue.IsStrSlice() {
//		value = tmpValue.StrSlice()
//	} else if tmpValue.IsUint() {
//		value = tmpValue.Uint()
//	} else if tmpValue.IsUintSlice() {
//		value = tmpValue.UintSlice()
//	} else if tmpValue.IsUint8() {
//		value = tmpValue.Uint8()
//	} else if tmpValue.IsUint8Slice() {
//		value = tmpValue.Uint8Slice()
//	} else if tmpValue.IsUint16() {
//		value = tmpValue.Uint16()
//	} else if tmpValue.IsUint16Slice() {
//		value = tmpValue.Uint16Slice()
//	} else if tmpValue.IsUint32() {
//		value = tmpValue.Uint32()
//	} else if tmpValue.IsUint32Slice() {
//		value = tmpValue.Uint32Slice()
//	} else if tmpValue.IsUint64() {
//		value = tmpValue.Uint64()
//	} else if tmpValue.IsUint64Slice() {
//		value = tmpValue.Uint64Slice()
//	} else if tmpValue.IsUintptr() {
//		value = tmpValue.Uintptr()
//	} else if tmpValue.IsUintptrSlice() {
//		value = tmpValue.UintptrSlice()
//	} else if tmpValue.IsMSI() {
//		value = normalize(tmpValue.MSI())
//	} else if tmpValue.IsObjxMap() {
//		value = normalize(tmpValue.ObjxMap())
//	} else if tmpValue.IsMSISlice() {
//		value = normalize(tmpValue.MSISlice())
//	} else if tmpValue.IsObjxMapSlice() {
//		value = normalize(tmpValue.ObjxMapSlice())
//	} else if tmpValue.IsInterSlice() {
//		value = normalize(tmpValue.InterSlice())
//	} else if tmpValue.IsInter() {
//		value = tmpValue.Inter()
//	} else if tmpValue.IsNil() {
//		value = nil
//	}
//
//	return value, nil
//}

//func GetMap(src map[string]interface{}, key string) (map[string]interface{}, error) {
//	value, err := Get(src, key)
//	if err != nil {
//		return nil, err
//	}
//
//	return value.(map[string]interface{}), nil
//}

//func Set(src map[string]interface{}, key string, value interface{}, merge bool) (map[string]interface{}, error) {
//	tmpMap := objx.New(src)
//
//	if merge {
//		tmpMap.Merge(objx.New(makeMap(key, value)))
//	} else {
//		selector := makeSelector(key)
//		if selector == "" {
//			tmpMap = objx.New(value)
//		} else {
//			if tmpMap.Has(selector) {
//				tmpMap = tmpMap.Set(selector, objx.New(value))
//			} else {
//				tmpMap = tmpMap.Merge(objx.New(makeMap(key, value)))
//			}
//		}
//	}
//
//	src = normalize(tmpMap).(map[string]interface{})
//
//	return src, nil
//}

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
