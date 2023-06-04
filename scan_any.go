package ngorm

import "reflect"

func (e *entity) scanAny(model *Model) error {
	e.formatSet()
	if e.err != nil {
		return e.err
	}

	model.rv.Set(reflect.Indirect(reflect.ValueOf(e.formatted)))

	return nil
}
