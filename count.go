package ngorm

func (e *entity) Count(value ...*int64) error {
	e.execute(0)
	if e.err != nil {
		return e.err
	}
	return e.count(value...)
}

func (e *entity) count(value ...*int64) error {
	var (
		columns = e.set.GetColNames()
	)
	if len(columns) == 0 || e.set.GetRowSize() == 0 {
		return ErrResultNil
	}
	values := e.set.GetRows()[0].GetValues()
	for i := range value {
		if len(values) > i {
			*value[i] = *values[i].IVal
		}
	}
	return nil
}
