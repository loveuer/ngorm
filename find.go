package ngorm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

func (e *entity) FindPath(number string, dest any) error {
	e.execute(0)
	if e.err != nil {
		return e.err
	}

	model, err := parse(dest)
	if err != nil {
		return err
	}

	if !model.rv.CanSet() {
		return ErrScanObject
	}

	if model.isAny {
		return e.scanAny(model)
	}

	if model.isMap {
		return e.scanMap(model)
	}

	return e.findPath(number, model)
}

func (e *entity) findPath(number string, model *Model) error {
	colVal, _ := e.set.GetValuesByColName(number)
	for _, colRow := range colVal {
		if colRow.IsEdge() {
			rv := reflect.New(model.rt)
			if !rv.CanSet() {
				rv = rv.Elem()
			}
			reg := regexp.MustCompile(`([0-9A-Za-z]+)(?:"->")([0-9A-Za-z]+).+(\[.+"\])`)
			colRowslice := reg.FindStringSubmatch(colRow.String())
			if len(colRowslice) == 4 {
				if _, ok := model.Fields["names"]; ok {
					if err := json.Unmarshal([]byte(colRowslice[3]), rv.FieldByName(model.Fields["names"]).Addr().Interface()); err != nil {
						e.logger.Debug(fmt.Sprintf("colRow as list[as string] unmarshal to []any err: %v", err))
						return err
					}
				}
				if _, ok := model.Fields["src"]; ok {
					rv.FieldByName(model.Fields["src"]).SetString(colRowslice[1])
				}
				if _, ok := model.Fields["dst"]; ok {
					rv.FieldByName(model.Fields["dst"]).SetString(colRowslice[2])
				}
			}
			if _, ok := model.Fields["edge"]; ok {
				rv.FieldByName(model.Fields["edge"]).SetString(colRow.String())
			}
			if !model.isSlice {
				model.rv.Set(rv)
				return nil
			}
			if model.rv.Type().Elem().Kind() == reflect.Ptr {
				model.rv.Set(reflect.Append(model.rv, rv.Addr()))
			} else {
				model.rv.Set(reflect.Append(model.rv, rv))
			}

		}
	}
	return nil
}

func (e *entity) Finds(key string, values ...any) error {
	e.execute(0)
	if e.err != nil {
		return e.err
	}
	var (
		colNames = e.set.GetColNames()
	)
	if len(colNames) == 0 || e.set.GetRowSize() == 0 {
		return ErrResultNil
	}
	for i := range values {
		if len(colNames) > i {
			colVal, err := e.set.GetValuesByColName(colNames[i])
			if err != nil {
				return err
			}
			for _, colRow := range colVal {
				model, err := parse(values[i])
				if err != nil {
					return err
				}
				rv := reflect.New(model.rt)
				if !rv.CanSet() {
					rv = rv.Elem()
				}
				e.scanNodev2(key, colRow, rv, model)
				if !model.isSlice {
					model.rv.Set(rv)
					return nil
				}
				if model.rv.Type().Elem().Kind() == reflect.Ptr {
					model.rv.Set(reflect.Append(model.rv, rv.Addr()))
				} else {
					model.rv.Set(reflect.Append(model.rv, rv))
				}
			}
		}
	}
	return nil
}

func (e *entity) scanNodev2(key string, vw *nebula.ValueWrapper, rv reflect.Value, model *Model) error {
	switch vw.GetType() {
	case "null":
	case "bool":
		data, err := vw.AsBool()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(data))
	case "int":
		data, err := vw.AsInt()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(data))
	case "float":
		data, err := vw.AsFloat()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(data))
	case "string":
		data := vw.String()
		if rv.Type().Kind() == reflect.Array|reflect.Slice {
			var (
				bsStr string
				err   error
			)
			if bsStr, err = vw.AsString(); err != nil {
				e.logger.Debug(fmt.Sprintf("vw as list[as string] err: %v", err))
				return err
			}

			if err = json.Unmarshal([]byte(bsStr), rv.Addr().Interface()); err != nil {
				e.logger.Debug(fmt.Sprintf("vw as list[as string] unmarshal to []any err: %v", err))
				return err
			}
		} else {
			rv.Set(reflect.ValueOf(data))
		}
	case "date":
		data := vw.String() // Time HH:MM:SS.MSMSMS
		newTime, err := time.Parse("15:04:05.999999999", data)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(newTime))
	case "time":
		data := vw.String() // Date yyyy-mm-dd
		newTime, err := time.Parse("2006-01-02", data)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(newTime))
	case "datetime":
		data := vw.String() //yyyy-mm-ddTHH:MM:SS.MSMSMS
		newTime, err := time.Parse("2006-01-02T15:04:05.999999999", data)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(newTime))
	case "vertex":
		node, _ := vw.AsNode()
		if _, ok := model.Fields["VertexID"]; ok {
			id, _ := node.GetID().AsString()
			rv.FieldByName(model.Fields["VertexID"]).SetString(id)
		}
		for colName, field := range model.Fields {
			if colName == "VertexID" {
				continue
			}
			propMap, err := node.Properties(colName)
			if err != nil {
				continue
			}
			if key != "" {
				nVw, ok := propMap[key]
				if !ok {
					e.logger.Debug("not found vw")
				}
				nRv := rv.FieldByName(field)
				if nModel, err := parse(nRv.Addr().Interface()); err != nil {
					if err != nil {
						return err
					}
				} else {
					e.scanNodev2(key, nVw, nRv, nModel)
					continue
				}
			}
			colRv := rv.FieldByName(field)
			colModel, err := parse(colRv.Addr().Interface())
			if err != nil {
				return err
			}
			for k, nVw := range propMap {
				nRv := colRv.FieldByName(colModel.Fields[k])
				if nModel, err := parse(nRv.Addr().Interface()); err != nil {
					if err != nil {
						return err
					}
				} else {
					e.scanNodev2(key, nVw, nRv, nModel)
				}
			}

		}
	case "edge":

	case "path":

	case "list":

	case "map":
		data, err := vw.AsMap()
		if err != nil {
			return err
		}
		for field, nVw := range data {
			nRv := rv.FieldByName(model.Fields[field])
			if nModel, err := parse(nRv.Addr().Interface()); err != nil {
				if err != nil {
					return err
				}
			} else {
				e.scanNodev2(key, &nVw, nRv, nModel)
			}
		}
	case "set":

	case "geography":
		err := scanGeography(vw, rv)
		if err != nil {
			return err
		}
	case "duration":

	case "empty":

	}
	return nil
}
func scanGeography(vw *nebula.ValueWrapper, rv reflect.Value) error {
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.String:
		rv.Set(reflect.ValueOf(vw.String()))
	case reflect.Struct:
		data, err := vw.AsGeography()
		if err != nil {
			return err
		}
		if data.IsSetPtVal() {
			ptVal := data.GetPtVal()
			for idx := 0; idx < rt.NumField(); idx++ {
				tag := strings.TrimSpace(rt.Field(idx).Tag.Get("nebula"))
				if tag == "coord" {
					nRv := rv.FieldByName(rt.Field(idx).Name)
					for idx := 0; idx < nRv.Type().NumField(); idx++ {
						tag := strings.TrimSpace(rt.Field(idx).Tag.Get("nebula"))
						if tag == "x" {
							nRv.Set(reflect.ValueOf(ptVal.Coord.X))
						} else if tag == "y" {
							nRv.Set(reflect.ValueOf(ptVal.Coord.Y))
						}
					}
					return nil
				}
			}
		} else if data.IsSetLsVal() {

		} else if data.IsSetPgVal() {

		}

	}
	return nil
}
