package ngorm

import (
	"log"
	"reflect"
	"strings"
)

type Tag struct {
	Index    int
	TagValue string
	TagKey   string
}

type Model struct {
	isAny   bool
	isSlice bool
	isMap   bool
	tags    map[int]*Tag
	rv      reflect.Value
	rt      reflect.Type
}

func Parse(dest any) (*Model, error) {
	var (
		rv    = reflect.ValueOf(dest)
		rt    = reflect.TypeOf(dest)
		model = &Model{tags: make(map[int]*Tag)}
	)

	defer func() {
		log.Printf("[reflect] type kind: %v", model.rt.Kind())
	}()

	if rv.Type().Kind() != reflect.Ptr {
		return nil, ErrNotPtr
	}

	rv = rv.Elem()
	rt = rt.Elem()

	if rt.Kind() == reflect.Array || rt.Kind() == reflect.Slice {
		model.isSlice = true
	}

	rt = findReflectType(rt)

	model.rv = rv
	model.rt = rt

	switch model.rt.Kind() {
	case reflect.Interface:
		model.isAny = true
	case reflect.Map:
		model.isMap = true
	case reflect.Struct:
		model.tags = findModelTags(model.rt)
	}

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

func findModelTags(rt reflect.Type) map[int]*Tag {
	tagMap := make(map[int]*Tag)
	for idx := 0; idx < rt.NumField(); idx++ {
		tag := strings.TrimSpace(rt.Field(idx).Tag.Get("nebula"))
		if tag == "" || tag == "-" {
			continue
		}

		vs := strings.Split(tag, ".")
		if len(vs) >= 2 {
			tagMap[idx] = &Tag{Index: idx, TagValue: vs[0], TagKey: vs[1]}
		} else {
			tagMap[idx] = &Tag{Index: idx, TagValue: vs[0], TagKey: "v"}
		}
	}

	return tagMap
}
