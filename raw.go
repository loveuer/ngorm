package ngorm

import nebula "github.com/vesoft-inc/nebula-go/v2"

type RawController struct {
	db   *NGDB
	tag  string
	ngql string
	err  error
}

func (db *NGDB) Raw(ngql string) *RawController {
	var (
		rc = &RawController{db: db, ngql: ngql}
	)

	return rc
}

func (r *RawController) genngql() (string, error) {
	return r.ngql, nil
}

func (r *RawController) Value() (*nebula.ResultSet, error) {
	e := &entry{db: r.db, ctrl: r}

	return e.value()
}
