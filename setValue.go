package ngorm

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

func getFieldMapIdx(rtype reflect.Type) map[string]int {
	m := make(map[string]int)
	for index := 0; index < rtype.NumField(); index++ {
		ftag := rtype.Field(index).Tag.Get("nebula")
		if tag := strings.TrimSpace(ftag); tag != "" && tag != "-" {
			m[tag] = index
			logrus.Infof("nebula tag: [tag: %s] [model: %s]", tag, rtype.Kind().String())
		}
	}

	return m
}

type modelType struct {
	isArray  bool
	isStruct bool
	isMap    bool
	fm       map[string]int
	rtype    reflect.Type
	rvalue   reflect.Value
}

func parseModel(model interface{}) (modelType, error) {
	var (
		mt = modelType{}
	)

	mt.rvalue = reflect.ValueOf(model)
	if mt.rvalue.Type().Kind() != reflect.Ptr {
		logrus.Errorf("model type not ptr, but: %s", mt.rvalue.Type().String())
		return mt, ErrorModelNotPtr
	}

	mt.rvalue = mt.rvalue.Elem()

	if !mt.rvalue.IsValid() {
		return mt, ErrorInvalidModel
	}

	if mt.rvalue.Type().Kind() == reflect.Slice || mt.rvalue.Type().Kind() == reflect.Array {
		mt.isArray = true
		mt.rtype = mt.rvalue.Type().Elem()
	} else {
		mt.rtype = mt.rvalue.Type()
	}

	if mt.rtype.Kind() == reflect.Struct {
		mt.isStruct = true
		mt.fm = getFieldMapIdx(mt.rtype)
	} else if mt.rtype.Kind() == reflect.Map {
		mt.isMap = true
	}

	pt := mt.rtype.Kind().String()
	if mt.isArray {
		pt = "[]" + pt
	}
	logrus.Infof("model type: %s", pt)

	return mt, nil
}

func bindValueByRow(val *nebula.ResultSet, model interface{}) error {
	var (
		err error
		mt  modelType
	)

	if mt, err = parseModel(model); err != nil {
		return err
	}

	_ = mt

	rowSize := val.GetRowSize()
	if rowSize == 0 {
		return ErrorResultNotFound
	}

	colSize := val.GetColSize()

	for idx := 0; idx < rowSize; idx++ {
		rowVal, err := val.GetRowValuesByIndex(idx)
		if err != nil {
			logrus.Errorf("get row by index err, idx: %d, err: %v", idx, err)
			return err
		}

		for colidx := 0; colidx < colSize; colidx++ {
			valWrapper, err := rowVal.GetValueByIndex(colidx)
			if err != nil {
				logrus.Errorf("get col value by idx err, idx: %d, err: %v", colidx, err)
				return err
			}

			logrus.Info("val wrapper:", valWrapper.GetType())
		}
	}

	return nil
}
