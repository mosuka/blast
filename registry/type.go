package registry

import (
	"errors"
	"fmt"
	"reflect"
)

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
	switch ins := instance.(type) {
	case map[string]interface{}:
		return reflect.TypeOf(ins).String()
	default:
		return reflect.TypeOf(ins).Elem().String()
	}
}

func TypeInstanceByName(name string) interface{} {
	return reflect.New(TypeByName(name)).Interface()
}
