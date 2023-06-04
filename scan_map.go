package ngorm

import "reflect"

func (e *entity) scanMap(model *Model) error {
	e.formatSet()
	if e.err != nil {
		return e.err
	}

	if !model.isSlice {
		if len(e.formatted) > 0 {
			model.rv.Set(reflect.Indirect(reflect.ValueOf(e.formatted[0])))
		}
	} else {
		model.rv.Set(reflect.Indirect(reflect.ValueOf(e.formatted)))
	}

	return nil
}
