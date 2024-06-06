package ngorm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-go/v3/nebula"
)

func (e *entity) Finds(key string, values ...any) error {
	e.execute()
	if e.err != nil {
		return e.err
	}
	var (
		colNames = e.set.GetColNames()
	)
	if len(colNames) == 0 || e.set.GetRowSize() == 0 {
		return ErrResultNil
	}
	for _, row := range e.set.GetRows() {
		for i2, value := range row.GetValues() {
			if len(values) > i2 {
				model, err := parse(values[i2])
				if err != nil {
					return err
				}
				rv := reflect.New(model.rt)
				if !rv.CanSet() {
					rv = rv.Elem()
				}
				e.scanNebulaValue(key, value, rv, model)
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

func scanGeographyToValue(vw *nebula.Value, rv reflect.Value) error {
	data := vw.GetGgVal()
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.String:
		rv.Set(reflect.ValueOf(data.String()))
	case reflect.Struct:
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
func (e *entity) scanEdge(key string, nValue *nebula.Value, rv reflect.Value, model *Model) error {
	edge := nValue.GetEVal()
	fmt.Println("string(edge.Src.GetSVal()):", string(edge.Src.GetSVal()))
	if _, ok := model.Fields["src"]; ok {
		rv.FieldByName(model.Fields["src"]).SetString(string(edge.Src.GetSVal()))
	}
	if _, ok := model.Fields["dst"]; ok {
		rv.FieldByName(model.Fields["dst"]).SetString(string(edge.Dst.GetSVal()))
	}
	if _, ok := model.Fields["rank"]; ok {
		rv.FieldByName(model.Fields["rank"]).SetInt(edge.GetRanking())
	}
	props := edge.Props
	if key != "" {
		if _, ok := model.Fields["names"]; ok {
			if err := json.Unmarshal(props[key].GetSVal(), rv.FieldByName(model.Fields["names"]).Addr().Interface()); err != nil {
				e.sess.logger.Debug("colRow as list[as string] unmarshal to []any err: %v", err)
				return err
			}
		}
	}
	return nil
}

func (e *entity) scanNebulaValue(key string, vw *nebula.Value, rv reflect.Value, model *Model) error {
	if vw.IsSetBVal() { //bool
		rv.Set(reflect.ValueOf(vw.GetBVal()))
	} else if vw.IsSetIVal() { //int
		rv.Set(reflect.ValueOf(vw.GetIVal()))
	} else if vw.IsSetFVal() { //float64
		rv.Set(reflect.ValueOf(vw.GetFVal()))
	} else if vw.IsSetSVal() { //string
		if rv.Type().Kind() == reflect.Array|reflect.Slice {
			if err := json.Unmarshal(vw.GetSVal(), rv.Addr().Interface()); err != nil {
				e.sess.logger.Debug("vw as list[as string] unmarshal to []any err: %v", err)
				return err
			}
		} else {
			rv.Set(reflect.ValueOf(string(vw.GetSVal())))
		}
	} else if vw.IsSetDVal() { //data
		data := vw.GetDVal() //
		newTime := time.Date(int(data.Year), time.Month(data.Month), int(data.Day), 0, 0, 0, 0, time.Local)
		rv.Set(reflect.ValueOf(newTime))
	} else if vw.IsSetDtVal() { //datetime
		data := vw.GetDtVal() //
		newTime := time.Date(int(data.Year), time.Month(data.Month), int(data.Day), int(data.Hour), int(data.Minute), int(data.Sec), int(data.Microsec), time.Local)
		rv.Set(reflect.ValueOf(newTime))
	} else if vw.IsSetTVal() { //time
		data := vw.GetTVal() //
		newTime := time.Date(0, 0, 0, int(data.Hour), int(data.Minute), int(data.Sec), int(data.Microsec), time.Local)
		rv.Set(reflect.ValueOf(newTime))
	} else if vw.IsSetEVal() { //edge
		e.scanEdge(key, vw, rv, model)
	} else if vw.IsSetVVal() { //Vertex
		vertex := vw.GetVVal()
		if _, ok := model.Fields["VertexID"]; ok {
			id := string(vertex.GetVid().GetSVal())
			rv.FieldByName(model.Fields["VertexID"]).SetString(id)
		}
		tags := vertex.GetTags()
		for _, tag := range tags {
			name := string(tag.Name)
			props := tag.GetProps()
			if _, ok := model.Fields[name]; ok {
				if key != "" {
					nVw := props[key]
					nRv := rv.FieldByName(model.Fields[name])
					if nModel, err := parse(nRv.Addr().Interface()); err != nil {
						if err != nil {
							return err
						}
					} else {
						e.scanNebulaValue(key, nVw, nRv, nModel)
						continue
					}
				}
				colRv := rv.FieldByName(model.Fields[name])
				colModel, err := parse(colRv.Addr().Interface())
				if err != nil {
					return err
				}
				for k, nVw := range props {
					nRv := colRv.FieldByName(colModel.Fields[k])
					if nModel, err := parse(nRv.Addr().Interface()); err != nil {
						if err != nil {
							return err
						}
					} else {
						e.scanNebulaValue(key, nVw, nRv, nModel)
					}
				}
			}

		}
	} else if vw.IsSetMVal() { //map
		data := vw.GetMVal()
		for field, nVw := range data.GetKvs() {
			nRv := rv.FieldByName(model.Fields[field])
			if nModel, err := parse(nRv.Addr().Interface()); err != nil {
				if err != nil {
					return err
				}
			} else {
				e.scanNebulaValue(key, nVw, nRv, nModel)
			}
		}
	} else if vw.IsSetLVal() { //list

	} else if vw.IsSetGgVal() { //geography
		err := scanGeographyToValue(vw, rv)
		if err != nil {
			return err
		}
	} else if vw.IsSetGVal() { //DataSet

	}
	return nil
}
