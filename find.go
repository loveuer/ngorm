package ngorm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
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
