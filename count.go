package ngorm

func (e *entity) Count(value *int64) error {
	e.execute(0)
	if e.err != nil {
		return e.err
	}
	return e.count(value)
}

func (e *entity) count(value *int64) error {
	var (
		columns = e.set.GetColNames()
	)
	if len(columns) == 0 || e.set.GetRowSize() == 0 {
		return ErrResultNil
	}
	row := e.set.GetRows()[0]
	if len(row.GetValues()) == 0 {
		return ErrResultNil
	}
	*value = *(row.GetValues()[0].IVal)
	return nil
}
