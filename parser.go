package ngorm

import (
	"reflect"
	"strings"
)

type Tag struct {
	Index    int
	TagValue string
}

type Model struct {
	isAny   bool
	isSlice bool
	isMap   bool
	tags    map[string]*Tag
	Fields  map[string]string
	rv      reflect.Value
	rt      reflect.Type
}

func parse(dest any) (*Model, error) {
	var (
		rv    = reflect.ValueOf(dest)
		rt    = reflect.TypeOf(dest)
		model = &Model{tags: make(map[string]*Tag)}
	)

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
		model.Fields = findModelFileds(model.rt)
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

func findModelTags(rt reflect.Type) map[string]*Tag {
	tagMap := make(map[string]*Tag)
	for idx := 0; idx < rt.NumField(); idx++ {
		tag := strings.TrimSpace(rt.Field(idx).Tag.Get("nebula"))
		if tag == "" || tag == "-" {
			continue
		}

		tagMap[tag] = &Tag{Index: idx, TagValue: tag}
	}

	return tagMap
}

func findModelFileds(rt reflect.Type) map[string]string {
	fields := make(map[string]string)
	for idx := 0; idx < rt.NumField(); idx++ {
		tag := strings.TrimSpace(rt.Field(idx).Tag.Get("nebula"))
		if tag == "" || tag == "-" {
			continue
		}
		fields[tag] = rt.Field(idx).Name
	}

	return fields
}
