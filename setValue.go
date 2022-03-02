package ngorm

import (
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
		log.Debugf("model type not ptr, but: %s", mt.rvalue.Type().String())
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
	log.Infof("model type: %s", pt)

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
			log.Debugf("get row by index err, idx: %d, err: %v", idx, err)
			return err
		}

		for colidx := 0; colidx < colSize; colidx++ {
			valWrapper, err := rowVal.GetValueByIndex(colidx)
			if err != nil {
				log.Debugf("get col value by idx err, idx: %d, err: %v", colidx, err)
				return err
			}

			log.Info("val wrapper:", valWrapper.GetType())
		}
	}

	return nil
}

func appendValue(val *nebula.ValueWrapper, model interface{}, mt modelType) error {
	return nil
}

// func setValue(set *nebula.ResultSet, models ...interface{}) error {
// 	if set.GetRowSize() < 1 {
// 		return ErrorResultNotFound
// 	}

// 	if len(models) != set.GetColSize() {
// 		return errors.New("models not fit with returns")
// 	}

// 	var (
// 		mts = make([]mType, 0, len(models))
// 	)

// 	for _, model := range models {
// 		mt, err := parseModel(model)
// 		if err != nil {
// 			return err
// 		}
// 		mts = append(mts, mt)
// 	}

// 	for ridx := 0; ridx < set.GetRowSize(); ridx++ {
// 		row, err := set.GetRowValuesByIndex(ridx)
// 		if err != nil {
// 			if Debug {

// 			}
// 			continue
// 		}

// 		for cidx := 0; cidx < set.GetColSize(); cidx++ {
// 			if !mts[cidx].isArray && ridx > 0 {
// 				continue
// 			}

// 			col, err := row.GetValueByIndex(cidx)
// 			if err != nil {
// 				if Debug {

// 				}
// 				continue
// 			}

// 			switch {
// 			case col.IsVertex():
// 				val, err := col.AsNode()
// 				if err != nil {
// 					log.Debugf("column value as vertex err: %v", err)
// 					continue
// 				}

// 				setValueNode(val, mts[cidx])
// 			case col.IsEdge():
// 				val, err := col.AsRelationship()
// 				if err != nil {
// 					log.Debugf("column value as edge err: %v", err)
// 					continue
// 				}

// 				setValueRelationship(val, mts[cidx])
// 			case col.IsString():
// 				val, err := col.AsString()
// 				if err != nil {
// 					log.Debugf("column value as str err: %v", err)
// 					continue
// 				}
// 				mt := mts[cidx]
// 				if mt.rtype.Kind() != reflect.String {
// 					return ErrorInvalidModelType
// 				}
// 				setValueNormal(val, mt)
// 			case col.IsFloat():
// 				val, err := col.AsFloat()
// 				if err != nil {
// 					log.Debugf("column value as float err: %v", err)
// 					continue
// 				}
// 				mt := mts[cidx]
// 				if mt.rtype.Kind() != reflect.Float64 {
// 					return ErrorInvalidModelType
// 				}
// 				setValueNormal(val, mt)
// 			case col.IsInt():
// 				val, err := col.AsInt()
// 				if err != nil {
// 					log.Debugf("column value as int err: %v", err)
// 					continue
// 				}
// 				mt := mts[cidx]
// 				switch mt.rtype.Kind() {
// 				case reflect.Int64:
// 					setValueNormal(val, mt)
// 				case reflect.Int32:
// 					setValueNormal(int32(val), mt)
// 				case reflect.Int16:
// 					setValueNormal(int16(val), mt)
// 				case reflect.Int8:
// 					setValueNormal(int8(val), mt)
// 				case reflect.Uint64:
// 					setValueNormal(uint64(val), mt)
// 				case reflect.Uint32:
// 					setValueNormal(uint32(val), mt)
// 				case reflect.Uint16:
// 					setValueNormal(uint16(val), mt)
// 				case reflect.Uint8:
// 					setValueNormal(uint8(val), mt)
// 				case reflect.Int:
// 					setValueNormal(int(val), mt)
// 				case reflect.Uint:
// 					setValueNormal(uint(val), mt)
// 				default:
// 					return ErrorInvalidModelType
// 				}
// 			case col.IsBool():
// 				val, err := col.AsBool()
// 				if err != nil {
// 					log.Debugf("column value as bool err: %v", err)
// 					continue
// 				}
// 				mt := mts[cidx]
// 				setValueNormal(val, mt)
// 			case col.IsNull():
// 				mt := mts[cidx]
// 				setValueNormal(reflect.Zero(mt.rtype), mt)
// 			default:
// 				return errors.New("value type not supported")
// 			}
// 		}
// 	}

// 	return nil
// }

// func setValueNormal(val interface{}, mt mType) {
// 	if mt.isArray {
// 		mt.rvalue.Set(reflect.Append(mt.rvalue, reflect.ValueOf(val)))
// 	} else {
// 		mt.rvalue.Set(reflect.ValueOf(val))
// 	}
// }

// func setValueNode(val *nebula.Node, mt mType) {
// 	tags := val.GetTags()
// 	if len(tags) < 1 {
// 		return
// 	}

// 	props, err := val.Properties(tags[0])
// 	if err != nil {
// 		log.Warnf("get props err, tags: %v,  model_type: %+v, ", tags, mt)
// 		return
// 	}

// 	vid := val.GetID()

// 	switch {
// 	case mt.isMap:
// 		m := make(map[string]interface{})

// 		switch {
// 		case vid.IsInt():
// 			vv, _ := vid.AsInt()
// 			m["vid"] = vv
// 		case vid.IsString():
// 			vv, _ := vid.AsString()
// 			m["vid"] = vv
// 		case vid.IsFloat():
// 			vv, _ := vid.AsFloat()
// 			m["vid"] = vv
// 		}

// 		for pk, pv := range props {
// 			m[pk] = pv
// 		}

// 		if !mt.isArray {
// 			mt.rvalue.Set(reflect.ValueOf(m))
// 			return
// 		}

// 		mt.rvalue.Set(reflect.Append(mt.rvalue, reflect.ValueOf(m)))
// 	case mt.isStruct:
// 		one := reflect.New(mt.rtype)

// 		if vidx, ok := mt.fm["vid"]; ok {
// 			switch {
// 			case vid.IsInt():
// 				vv, _ := vid.AsInt()
// 				one.Elem().Field(vidx).SetInt(vv)
// 			case vid.IsString():
// 				vv, _ := vid.AsString()
// 				one.Elem().Field(vidx).SetString(vv)
// 			case vid.IsFloat():
// 				vv, _ := vid.AsFloat()
// 				one.Elem().Field(vidx).SetFloat(vv)
// 			}
// 		}

// 		for pk, pv := range props {
// 			midx, ok := mt.fm[pk]
// 			if ok {
// 				switch {
// 				case pv.IsInt():
// 					val, err := pv.AsInt()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetInt(val)
// 				case pv.IsString():
// 					val, err := pv.AsString()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetString(val)
// 				case pv.IsBool():
// 					val, err := pv.AsBool()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetBool(val)
// 				case pv.IsFloat():
// 					val, err := pv.AsFloat()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetFloat(val)
// 				case pv.IsNull():
// 					one.Elem().Field(midx).Set(reflect.ValueOf(nil))
// 				}
// 			}
// 		}

// 		if !mt.isArray {
// 			mt.rvalue.Set(reflect.Indirect(one))
// 			return
// 		}

// 		mt.rvalue.Set(reflect.Append(mt.rvalue, reflect.Indirect(one)))
// 	}
// }

// func setValueRelationship(val *nebula.Relationship, mt mType) {
// 	props := val.Properties()
// 	switch {
// 	case mt.isMap:
// 		m := make(map[string]interface{})
// 		for pk, pv := range props {
// 			m[pk] = pv
// 		}

// 		if !mt.isArray {
// 			mt.rvalue.Set(reflect.ValueOf(m))
// 			return
// 		}

// 		mt.rvalue.Set(reflect.Append(mt.rvalue, reflect.ValueOf(m)))
// 	case mt.isStruct:
// 		one := reflect.New(mt.rtype)
// 		for pk, pv := range props {
// 			midx, ok := mt.fm[pk]
// 			if ok {
// 				switch {
// 				case pv.IsInt():
// 					val, err := pv.AsInt()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetInt(val)
// 				case pv.IsString():
// 					val, err := pv.AsString()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetString(val)
// 				case pv.IsBool():
// 					val, err := pv.AsBool()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetBool(val)
// 				case pv.IsFloat():
// 					val, err := pv.AsFloat()
// 					if err != nil {
// 						continue
// 					}
// 					one.Elem().Field(midx).SetFloat(val)
// 				case pv.IsNull():
// 					one.Elem().Field(midx).Set(reflect.ValueOf(nil))
// 				}
// 			}
// 		}

// 		if !mt.isArray {
// 			mt.rvalue.Set(reflect.Indirect(one))
// 			return
// 		}

// 		mt.rvalue.Set(reflect.Append(mt.rvalue, reflect.Indirect(one)))
// 	}
// }

// func appendValueNode(tag string, node *nebula.Node, rvalue reflect.Value) {
// 	var (
// 		array bool
// 		rtype reflect.Type
// 	)

// 	if rvalue.Type().Kind() == reflect.Array || rvalue.Type().Kind() == reflect.Slice {
// 		array = true
// 		rtype = rvalue.Type().Elem()
// 	} else {
// 		array = false
// 		rtype = rvalue.Type()
// 	}
// 	props, err := node.Properties(tag)
// 	if err != nil {
// 		log.Warnf("node props not found by tag: %s", tag)
// 		return
// 	}

// 	for _, pv := range props {
// 		var (
// 			val interface{}
// 			err error
// 		)
// 		switch {
// 		case pv.IsInt():
// 			if !strings.Contains(strings.ToLower(rtype.Kind().String()), "int") {
// 				continue
// 			}
// 			val, err = pv.AsInt()
// 			if err != nil {
// 				continue
// 			}
// 		case pv.IsString() && rtype.Kind() == reflect.String:
// 			val, err = pv.AsString()
// 			if err != nil {
// 				continue
// 			}
// 		case pv.IsBool() && rtype.Kind() == reflect.Bool:
// 			val, err = pv.AsBool()
// 			if err != nil {
// 				continue
// 			}
// 		case pv.IsFloat() && (rtype.Kind() == reflect.Float32 || rtype.Kind() == reflect.Float64):
// 			val, err = pv.AsFloat()
// 			if err != nil {
// 				continue
// 			}
// 		case pv.IsNull():
// 			val = nil
// 		}

// 		rval := reflect.ValueOf(val)
// 		if rval.IsValid() {
// 			if array {
// 				rvalue.Set(reflect.Append(rvalue, reflect.ValueOf(val)))
// 			} else {
// 				rvalue.Set(reflect.ValueOf(val))
// 			}
// 		}

// 	}

// 	return
// }

// func appendValueRelationship(path *nebula.Relationship, mType mType, model interface{}) {

// 	vmap := map[string]string{
// 		"from": path.GetSrcVertexID().String(),
// 		"to":   path.GetDstVertexID().String(),
// 		"type": path.GetEdgeName(),
// 	}

// 	one := reflect.New(mType.rtype)

// 	for key, val := range vmap {

// 		if fidx, ok := mType.fm[key]; ok {
// 			rval := one.Elem().Field(fidx)
// 			rval.Set(reflect.ValueOf(val))
// 		}
// 	}

// 	mType.rvalue.Set(reflect.Append(mType.rvalue, reflect.Indirect(one)))
// }
