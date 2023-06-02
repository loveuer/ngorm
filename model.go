package ngorm

import (
	"log"
	"reflect"
)

type Model struct {
	isAny   bool
	isSlice bool
	isMap   bool
	rv      reflect.Value
	rt      reflect.Type
}

func Parse(dest any) (*Model, error) {
	var (
		rv    = reflect.ValueOf(dest)
		rt    = reflect.TypeOf(dest)
		model = &Model{}
	)

	defer func() {
		log.Printf("[reflect] type kind: %v", model.rt.Kind())
	}()

	if rv.Type().Kind() != reflect.Ptr {
		return nil, ErrNotPtr
	}

	rv = rv.Elem()
	rt = findReflectType(rt)

	model.rv = rv
	model.rt = rt

	return model, nil
}

func findReflectType(rt reflect.Type) reflect.Type {
	switch rt.Kind() {
	case reflect.Ptr:
		return findReflectType(rt.Elem())
	case reflect.Slice, reflect.Array:
		return findReflectType(rt.Elem())
	default:
		return rt
	}
}
