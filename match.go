package ngorm

import (
	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type MatchController struct {
	db              *NGDB
	tag             string
	ngql            string
	err             error
	sesseionTimeout int
}

func (db *NGDB) Match(ngql string) *MatchController {
	var (
		mc = &MatchController{db: db, ngql: ngql}
	)

	return mc
}

func (m *MatchController) SetTimeout(second int) *MatchController {
	m.sesseionTimeout = second
	return m
}

func (m *MatchController) genngql() (string, error) {
	return m.ngql, nil
}

// Value return the original nebula result
func (m *MatchController) Value() (*nebula.ResultSet, error) {

	e := &entry{db: m.db, ctrl: m, sessionTimeout: m.sesseionTimeout}

	return e.value()
}

func (m *MatchController) Finds(models ...interface{}) error {

	e := &entry{db: m.db, ctrl: m, sessionTimeout: m.sesseionTimeout}

	return e.finds(models...)
}
