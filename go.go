package ngorm

import (
	"errors"
	"fmt"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type GoController struct {
	db     *NGDB
	from   string
	steps  []int
	over   string
	yields []string
	err    error
}

func (db *NGDB) GOFrom(from string) *GoController {
	return &GoController{
		db:   db,
		from: from,
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

func (g *GoController) Yield(yield string) *GoController {
	g.yields = append(g.yields, yield)
	return g
}

func (g *GoController) genngql() (ngql string, err error) {
	if g.from == "" {
		return ngql, errors.New("from is ''")
	}

	if g.over == "" {
		return ngql, errors.New("over is ''")
	}

	if len(g.steps) > 2 {
		return ngql, errors.New("steps lenght only accept 0,1,2")
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

	ngql = fmt.Sprintf("GO %s FROM '%s' OVER %s %s", stepPart, g.from, g.over, yieldPart)

	return
}

func (g *GoController) Value() (*nebula.ResultSet, error) {
	e := &entry{db: g.db, ctrl: g}
	return e.value()
}

func (g *GoController) Find(model interface{}) error {
	e := &entry{db: g.db, ctrl: g}

	return e.find(model)
}
