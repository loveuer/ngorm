package ngorm

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"reflect"
	"time"
)

func (e *entity) scanValueWrapper(vw *nebula.ValueWrapper, column string, rv reflect.Value, model *Model) error {

	if vw.IsEmpty() || vw.IsNull() {
		rv.Set(reflect.New(rv.Type()).Elem())
		return nil
	}

	switch rv.Type().Kind() {
	case reflect.Invalid:
		return ErrModelReflectInvalid
	case reflect.Bool:
		val, err := vw.AsBool()
		if err != nil {
			e.logger.Debug(fmt.Sprintf("vw as bool err: %v", err))
			return err
		}

		rv.Set(reflect.ValueOf(val))

		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := vw.AsInt()
		if err != nil {
			e.logger.Debug(fmt.Sprintf("vw as int err: %v", err))
			return err
		}

		rv.Set(reflect.ValueOf(val))

		return nil
	case reflect.Float32, reflect.Float64:
		val, err := vw.AsFloat()
		if err != nil {
			e.logger.Debug(fmt.Sprintf("vw as float err: %v", err))
			return err
		}

		rv.Set(reflect.ValueOf(val))

		return nil
	case reflect.String:
		val, err := vw.AsString()
		if err != nil {
			e.logger.Debug(fmt.Sprintf("vw as string err: %v", err))
			return err
		}

		rv.SetString(val)

		return nil
	case reflect.Interface, reflect.Map:
		return fmt.Errorf("%w: todo: error model type detector", ErrModelTypeUnsupport)
	case reflect.Array, reflect.Slice:
		vws, err := vw.AsList()
		if err != nil {

			if vw.IsEmpty() {
				return nil
			}

			// 兼容 序列化的 结果
			var (
				bsStr string
				//is    = make([]any, 0)
			)
			if bsStr, err = vw.AsString(); err != nil {
				e.logger.Debug(fmt.Sprintf("vw as list[as string] err: %v", err))
				return err
			}

			if err = json.Unmarshal([]byte(bsStr), rv.Addr().Interface()); err != nil {
				e.logger.Debug(fmt.Sprintf("vw as list[as string] unmarshal to []any err: %v", err))
				return err
			}

			return nil
		}

		_ = vws

		// todo
		return fmt.Errorf("%w: todo: model type too complex", ErrModelTypeUnsupport)
	case reflect.Struct:
		tag := model.tags[column]
		if tag == nil {
			return nil
		}

		return e.scanValueWrapper(vw, column, rv.Field(tag.Index), model)
	case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		return fmt.Errorf("%w: %s", ErrModelTypeUnsupport, rv.Type().Kind().String())
	}

	return fmt.Errorf("%w: %s", ErrModelTypeUnsupport, rv.Type().Kind().String())
}

func (e *entity) handleValueWrapper(vw *nebula.ValueWrapper) (any, error) {
	switch {
	case vw.IsBool():
		val, _ := vw.AsBool()
		return val, nil
	case vw.IsFloat():
		val, _ := vw.AsFloat()
		return val, nil
	case vw.IsString():
		val, _ := vw.AsString()
		var parsed any
		if err := json.Unmarshal([]byte(val), &parsed); err == nil {
			return parsed, nil
		}
		return val, nil
	case vw.IsInt():
		val, _ := vw.AsInt()
		return val, nil
	case vw.IsDate():
		val, _ := vw.AsDate()
		t := time.Date(
			int(val.GetYear()),
			time.Month(val.GetMonth()),
			int(val.GetDay()),
			0, 0, 0, 0,
			time.Local,
		)
		return t, nil
	case vw.IsDateTime():
		var (
			val, _ = vw.AsDateTime()
		)

		d, err := val.GetLocalDateTimeWithTimezoneName(time.Local.String())
		if err != nil {
			e.logger.Debug(fmt.Sprintf("[ngorm] datetime value get local datetime err: %v", err))
			return nil, err
		}

		t := time.Date(
			int(d.GetYear()),
			time.Month(d.GetMonth()),
			int(d.GetDay()),
			int(d.GetHour()),
			int(d.GetMinute()),
			int(d.GetSec()),
			int(d.GetMicrosec()*1000),
			time.Local,
		)

		return t, nil
	case vw.IsDuration():
		val, _ := vw.AsDuration()
		d := time.Duration(val.GetSeconds()) * time.Second
		return d, nil
	case vw.IsEdge():
		// todo
		val, _ := vw.AsRelationship()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsEdge]")
	case vw.IsEmpty():
		return nil, nil
	case vw.IsGeography():
		// todo
		val, _ := vw.AsGeography()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsGeography]")
	case vw.IsList():
		val, _ := vw.AsList()
		list := make([]any, 0, len(val))

		for _, v := range val {
			data, err := e.handleValueWrapper(&v)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		return list, nil
	case vw.IsMap():
		val, _ := vw.AsMap()
		m := make(map[string]any, len(val))

		for k := range val {
			d := val[k]
			data, err := e.handleValueWrapper(&d)
			if err != nil {
				return nil, err
			}

			m[k] = data
		}

		return m, nil
	case vw.IsNull():
		return nil, nil
	case vw.IsPath():
		// todo
		val, _ := vw.AsPath()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsPath]")
	case vw.IsSet():
		val, _ := vw.AsDedupList()
		list := make([]any, 0, len(val))

		for _, v := range val {
			data, err := e.handleValueWrapper(&v)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		return list, nil
	case vw.IsTime():
		// todo
		val, _ := vw.AsTime()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsTime]")
	case vw.IsVertex():
		val, _ := vw.AsNode()
		tags := val.GetTags()
		m := make(map[string]map[string]any, len(tags))
		for _, tag := range tags {
			vwm, err := val.Properties(tag)

			if err != nil {
				e.logger.Debug(fmt.Sprintf("[ngorm] get properties by tag: %s, err: %v", tag, err))
				return nil, err
			}

			mt := make(map[string]any, len(vwm))

			for k := range vwm {
				data, err := e.handleValueWrapper(vwm[k])
				if err != nil {
					return nil, err
				}

				mt[k] = data
			}

			m[tag] = mt
		}

		return m, nil
	}

	return nil, ErrUnknownValueWrapper
}
