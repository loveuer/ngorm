package ngorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/cast"
	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type controller interface {
	genngql() (string, error)
}

type entry struct {
	db      *NGDB
	session *nebula.Session

	ctrl controller
}

func (e *entry) value() (*nebula.ResultSet, error) {
	var (
		set  *nebula.ResultSet
		ngql string
		err  error
	)

	if e.db == nil {
		return set, errors.New("entry with nil db")
	}

	ngql, err = e.ctrl.genngql()
	if err != nil {
		log.Debugf("generate ngql err: %v", err)
		return set, err
	}

	if ngql == "" {
		return set, errors.New("empty ngql")
	}

	e.session, err = e.db.prepare()
	if err != nil && Debug {
		log.Errorf("get session err: %v", err)
	}

	defer e.db.release(e.session)

	log.Infof("ngql: %s", ngql)

	set, err = e.session.Execute(ngql)
	if err != nil {
		log.Debug("session execute err:", err)
	}

	return set, err
}

func (e *entry) find(model interface{}) error {
	var (
		err       error
		ok        bool
		resultSet *nebula.ResultSet
		rowSize   int
		colMap    = make(map[string]struct{})
	)

	mt, err := parseModel(model)
	if err != nil {
		return err
	}

	if resultSet, err = e.value(); err != nil {
		return err
	}

	if rowSize = resultSet.GetRowSize(); rowSize == 0 {
		return ErrorResultNotFound
	}

	colNames := resultSet.GetColNames()
	for _, colName := range colNames {
		colMap[colName] = struct{}{}
	}

	// 验证 model 的 field 是否都在返回值有对应
	for key := range mt.fm {
		if _, ok = colMap[key]; !ok {
			return ErrorColumnNotFound(key)
		}
	}

	if !mt.isArray {
		row, err := resultSet.GetRowValuesByIndex(0)
		if err != nil {
			log.Debugf("nebula get row value by index err, index: 0, err: %v", err)
			return err
		}

		if mt.isMap { // mt is map
			mv, err := setRow2Map(row, colNames)
			if err != nil {
				return err
			}

			mt.rvalue.Set(reflect.ValueOf(mv))

			return nil
		}

		return setRow2Struct(row, mt.rvalue, mt) // mt is struct

	} else { // model type is array|slice
		if mt.isMap { // model type is []map[string]interface{}
			result := make([]map[string]interface{}, 0)

			for rowIdx := 0; rowIdx < rowSize; rowIdx++ {
				row, err := resultSet.GetRowValuesByIndex(rowIdx)
				if err != nil {
					log.Debugf("nebula get row value by index err, index: %d, err: %v", rowIdx, err)
					return err
				}

				mv, err := setRow2Map(row, colNames)
				if err != nil {
					return err
				}

				result = append(result, mv)

			}

			mt.rvalue.Set(reflect.ValueOf(result))

			return nil

		} else { // mt is []struct

			for rowIdx := 0; rowIdx < rowSize; rowIdx++ {
				row, err := resultSet.GetRowValuesByIndex(rowIdx)
				if err != nil {
					log.Debugf("nebula get row value by index err, index: %d, err: %v", rowIdx, err)
					return err
				}

				newOne := reflect.New(mt.rtype)
				if err = setRow2Struct(row, newOne, mt); err != nil {
					return err
				}

				mt.rvalue.Set(reflect.Append(mt.rvalue, newOne.Elem()))
			}

			return nil
		}
	}
}

func (e *entry) finds(models ...interface{}) error {
	var (
		err       error
		resultSet *nebula.ResultSet
		rowSize   int
	)

	if resultSet, err = e.value(); err != nil {
		return err
	}

	if rowSize = resultSet.GetRowSize(); rowSize == 0 {
		return ErrorResultNotFound
	}

	colSize := resultSet.GetColSize()
	if len(models) != colSize {
		return errors.New("model size not compatible with result column size")
	}

	colNames := resultSet.GetColNames()

	log.Debugf("col size: %d, model size: %d, col names: %v", colSize, len(models), colNames)

	for colIdx, colName := range colNames {
		log.Debugf("col idx: %d, col name: %s", colIdx, colName)
		mt, err := parseModel(models[colIdx])
		if err != nil {
			return err
		}

		colVal, err := resultSet.GetValuesByColName(colName)
		if err != nil {
			log.Debugf("nebula get column value by col_name err, col_name: %s, err: %v", colName, err)
			return err
		}

		if !mt.isArray { // mt is not array or slice
			if mt.isMap { // mt is map
				mv, err := setCell2Map(colVal[0])
				if err != nil {
					return err
				}

				mt.rvalue.Set(reflect.ValueOf(mv))

			} else { // mt is struct
				if err = setCell2Struct(colVal[0], mt.rvalue, mt); err != nil {
					return err
				}
			}
		} else { // mt is array || slice
			if mt.isMap { // mt is []map
				mvs := make([]map[string]interface{}, 0)
				for _, colRow := range colVal {
					mv, err := setCell2Map(colRow)
					if err != nil {
						return err
					}
					mvs = append(mvs, mv)
				}

				mt.rvalue.Set(reflect.ValueOf(mvs))

			} else { // mt is []struct
				for _, colRow := range colVal {
					oneNew := reflect.New(mt.rtype)
					if err = setCell2Struct(colRow, oneNew, mt); err != nil {
						return err
					}

					mt.rvalue.Set(reflect.Append(mt.rvalue, oneNew.Elem()))

				}

			}
		}

	}

	return nil
}

func setRow2Map(row *nebula.Record, colNames []string) (map[string]interface{}, error) {
	mv := make(map[string]interface{})
	for _, colName := range colNames {
		valWrapper, err := row.GetValueByColName(colName)
		if err != nil {
			log.Debugf("nebula get val by col_name err, col_name: %s, err: %v", colName, err)
			return mv, err
		}

		switch {
		case valWrapper.IsString():
			str, _ := valWrapper.AsString()
			strs := make([]string, 0)
			if err = json.Unmarshal([]byte(str), &strs); err != nil {
				mv[colName] = str
			} else {
				mv[colName] = strs
			}
		case valWrapper.IsEmpty(), valWrapper.IsNull():
			continue
		default:
			return mv, ErrorUnknownNebulaValueType(valWrapper.GetType())
		}
	}

	return mv, nil
}

func setCell2Map(cell *nebula.ValueWrapper) (map[string]interface{}, error) {
	mv := make(map[string]interface{})

	switch {
	case cell.IsVertex():
		node, _ := cell.AsNode()
		tags := node.GetTags()
		if vertexID, err := node.GetID().AsString(); err == nil {
			mv["VertexID"] = vertexID
		}

		for _, tag := range tags {
			props, err := node.Properties(tag)
			if err != nil {
				return mv, fmt.Errorf("get node tag err, tag: %s, err: %v", tag, err)
			}

			for _, propVal := range props {
				switch {
				case propVal.IsString():
					str, _ := propVal.AsString()
					strs := make([]string, 0)
					if err = json.Unmarshal([]byte(str), &strs); err != nil {
						mv[tag] = str
					} else {
						mv[tag] = strs
					}
				case propVal.IsEmpty(), propVal.IsNull():
					continue
				default:
					return mv, ErrorUnknownNebulaValueType(propVal.GetType())
				}
			}
		}
	}

	return mv, nil
}

func setRow2Struct(row *nebula.Record, rvalue reflect.Value, mt modelType) error {
	for col, idx := range mt.fm {
		valWrapper, err := row.GetValueByColName(col)
		if err != nil {
			log.Debugf("row get value by col name err, col_name: %s, err: %v", col, err)
			return err
		}

		if rvalue.Type().Kind() == reflect.Ptr {
			rvalue = rvalue.Elem()
		}

		if !rvalue.IsValid() || !rvalue.CanSet() {
			return fmt.Errorf("invalid reflect value: %s", rvalue.Type().Kind().String())
		}

		switch {
		case valWrapper.IsString():
			str, _ := valWrapper.AsString()
			ft := rvalue.Field(idx)
			if err = setStr(str, ft); err != nil {
				return err
			}
		case valWrapper.IsEmpty(), valWrapper.IsNull():
			continue
		default:
			return ErrorUnknownNebulaValueType(valWrapper.GetType())
		}
	}

	return nil
}

func setCell2Struct(cell *nebula.ValueWrapper, rvalue reflect.Value, mt modelType) error {
	if rvalue.Type().Kind() == reflect.Ptr {
		rvalue = rvalue.Elem()
	}

	if !rvalue.IsValid() || !rvalue.CanSet() {
		return fmt.Errorf("invalid reflect value: %s", rvalue.Type().Kind().String())
	}

	switch {
	case cell.IsString():
		str, _ := cell.AsString()
		if rvalue.Type().Kind() != reflect.String {
			return fmt.Errorf("nebula data is string, but model is: %s", rvalue.Type().String())
		}

		rvalue.Set(reflect.ValueOf(str))
		return nil

	case cell.IsVertex():
		node, _ := cell.AsNode()
		tags := node.GetTags()
		tagMap := make(map[string]struct{})
		for _, tag := range tags {
			tagMap[tag] = struct{}{}
		}

		for field, fidx := range mt.fm {
			ft := rvalue.Field(fidx)

			if field == "VertexID" {
				vertexID, err := node.GetID().AsString()
				if err != nil {
					log.Warnf("get vertex id err: %v", err)
				}

				if err = setStr(vertexID, ft); err != nil {
					return err
				}

				continue
			}

			props, err := node.Properties(field)
			if err != nil {
				log.Warnf("props not found by tag: %s", field)
				continue
			}


			for _, propVal := range props {
				switch {
				case propVal.IsString():
					str, _ := propVal.AsString()
					if err = setStr(str, ft); err != nil {
						return err
					}
				case propVal.IsEmpty(), propVal.IsNull():
					continue
				default:
					return ErrorUnknownNebulaValueType(propVal.GetType())
				}
			}
		}
	default:
		return ErrorUnknownNebulaValueType(cell.GetType())
	}

	return nil
}

// 针对我们采用了 json 序列化来存取 复杂数据结构的情况
func setStr(val string, ft reflect.Value) error {

	log.Infof("field type: %s", ft.Type().String())

	switch ft.Type().Kind() {
	case reflect.String:
		ft.Set(reflect.ValueOf(val))
		return nil

	case reflect.Int:
		vint, err := cast.ToIntE(val)
		if err != nil {
			log.Debugf("nebula val to int err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vint))
		return nil

	case reflect.Int8:
		vint8, err := cast.ToInt8E(val)
		if err != nil {
			log.Debugf("nebula val to int8 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vint8))
		return nil

	case reflect.Int16:
		vint16, err := cast.ToInt16E(val)
		if err != nil {
			log.Debugf("nebula val to int16 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vint16))
		return nil

	case reflect.Int32:
		vint32, err := cast.ToInt32E(val)
		if err != nil {
			log.Debugf("nebula val to int32 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vint32))
		return nil

	case reflect.Int64:
		vint64, err := cast.ToInt64E(val)
		if err != nil {
			log.Debugf("nebula val to int64 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vint64))
		return nil

	case reflect.Uint:
		vuint, err := cast.ToUintE(val)
		if err != nil {
			log.Debugf("nebula val to uint err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vuint))
		return nil

	case reflect.Uint8:
		vuint8, err := cast.ToUint8E(val)
		if err != nil {
			log.Debugf("nebula val to uint8 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vuint8))
		return nil

	case reflect.Uint16:
		vuint16, err := cast.ToUint16E(val)
		if err != nil {
			log.Debugf("nebula val to uint16 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vuint16))
		return nil

	case reflect.Uint32:
		vuint32, err := cast.ToUint32E(val)
		if err != nil {
			log.Debugf("nebula val to uint32 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vuint32))
		return nil

	case reflect.Uint64:
		vuint64, err := cast.ToUint64E(val)
		if err != nil {
			log.Debugf("nebula val to uint64 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vuint64))
		return nil

	case reflect.Float32:
		vf32, err := cast.ToFloat32E(val)
		if err != nil {
			log.Debugf("nebula val to float32 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vf32))
		return nil

	case reflect.Float64:
		vf64, err := cast.ToFloat64E(val)
		if err != nil {
			log.Debugf("nebula val to float64 err: %v", err)
			return err
		}
		ft.Set(reflect.ValueOf(vf64))
		return nil

	default:
		newOne := reflect.New(ft.Type())
		if err := json.Unmarshal([]byte(val), newOne.Interface()); err != nil {
			log.Debugf("unmashal nebula str val to field err, field type: %s, err: %v", ft.Type().String(), err)
			return err
		}

		ft.Set(newOne.Elem())

		return nil
	}
}