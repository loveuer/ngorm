package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"reflect"
)

func (e *entity) Scan(dest any) error {
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

	return e.scan(model)
}

func (e *entity) scan(model *Model) error {
	var (
		err     error
		columns = e.set.GetColNames()
	)

	if len(columns) == 0 || e.set.GetRowSize() == 0 {
		return ErrResultNil
	}

	for rowIndex := 0; rowIndex < e.set.GetRowSize(); rowIndex++ {
		var record *nebula.Record

		if record, err = e.set.GetRowValuesByIndex(rowIndex); e.err != nil {
			e.logger.Debug(fmt.Sprintf("[ngorm] get row: %d value err: %v", rowIndex, err))
			return err
		}

		rv := reflect.New(model.rt)

		if !rv.CanSet() {
			rv = rv.Elem()
		}

		for idx, column := range columns {
			var (
				vw *nebula.ValueWrapper
			)

			if vw, err = record.GetValueByColName(column); err != nil {
				e.logger.Debug(fmt.Sprintf("[ngorm] get row: %d column: %s value err: %v", rowIndex, column, err))
				return err
			}

			if err = e.scanValueWrapper(vw, column, rv.Field(idx), model); err != nil {
				return err
			}
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

	return nil
}
