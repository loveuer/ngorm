package ngorm

import (
	"errors"
	"fmt"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type GoController struct {
	db     *NGDB
	from   []string
	steps  []int
	over   string
	yields []string
	err    error
}

func (db *NGDB) GOFrom(from ...string) *GoController {
	var froms = make([]string, 0, len(from))
	for _, f := range from {
		froms = append(froms, fmt.Sprintf("'%s'", f))
	}

	return &GoController{
		db:   db,
		from: froms,
	}
}

func (g *GoController) Step(steps ...int) *GoController {
	g.steps = steps
	return g
}

func (g *GoController) Over(over string) *GoController {
	g.over = over
	return g
}

func (g *GoController) Yield(yields ...string) *GoController {
	g.yields = append(g.yields, yields...)
	return g
}

func (g *GoController) genngql() (ngql string, err error) {
	if len(g.from) == 0 {
		return ngql, errors.New("must specify from vertex")
	}

	if g.over == "" {
		return ngql, errors.New("over is ''")
	}

	if len(g.steps) > 2 {
		return ngql, errors.New("steps length only accept 0,1,2")
	}

	var (
		stepPart  string
		yieldPart string
	)

	switch len(g.steps) {
	case 0:
		stepPart = ""
	case 1:
		stepPart = fmt.Sprintf("%d STEPS", g.steps[0])
	case 2:
		stepPart = fmt.Sprintf("%d TO %d STEPS", g.steps[0], g.steps[1])
	}

	if len(g.yields) > 0 {
		yieldPart = "YIELD " + strings.Join(g.yields, ", ")
	}

	ngql = fmt.Sprintf("GO %s FROM %s OVER %s %s", stepPart, strings.Join(g.from, ", "), g.over, yieldPart)

	return
}

func (g *GoController) Value() (*nebula.ResultSet, error) {
	e := &entry{db: g.db, ctrl: g}
	return e.value()
}

func (g *GoController) Finds(models ...interface{}) error {
	e := &entry{db: g.db, ctrl: g}

	return e.finds(models...)
}
