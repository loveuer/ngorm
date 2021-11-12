package ngorm

import (
	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type MatchController struct {
	db      *NGDB
	tag     string
	ngql    string
	err     error
}

func (db *NGDB) Match(ngql string) *MatchController {
	var (
		mc = &MatchController{db: db, ngql: ngql}
	)

	return mc
}

func (m *MatchController) genngql() (string, error) {
	return m.ngql, nil
}

// Value return the original nebula result
func (m *MatchController) Value() (*nebula.ResultSet, error) {

	e := &entry{db: m.db, ctrl: m}

	return e.value()
}

func (m *MatchController) Finds(models ...interface{}) error {

	e := &entry{db: m.db, ctrl: m}

	return e.finds(models...)
}
