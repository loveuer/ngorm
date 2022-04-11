package ngorm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type GoController struct {
	db            *NGDB
	from          []string
	steps         []int
	over          string
	yields        []string
	step_limits   []int
	limit, offset int
	err           error
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

/*
StepLimit:
	- (文档详见)[https://docs.nebula-graph.com.cn/3.0.2/3.ngql-guide/8.clauses-and-options/limit/]
	- `限制每一步（steps）的 limits`
Examples:
	// 表示 1 - 3 步 分别限制 2个， 3个， 4个 结果
	// 在 管道 之前的限制，效果更加理想
	GOFrom("xxx").Step(1, 3).Over("edge").Yield("edge._dst").StepLimit([]int{2,3,4})
*/
func (g *GoController) StepLimit(limits []int) *GoController {
	g.step_limits = limits
	return g
}

func (g *GoController) Limit(limit int) *GoController {
	g.limit = limit
	return g
}

func (g *GoController) Offset(offset int) *GoController {
	g.offset = offset
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
		stepPart      string
		yieldPart     string
		stepLimits    string
		pipeLimitPart string
	)

	switch len(g.steps) {
	case 0:
		stepPart = ""
		if len(g.step_limits) > 1 {
			return ngql, errors.New("SemanticError: length must be equal to GO step size")
		} else if len(g.step_limits) == 1 {
			stepLimits = fmt.Sprintf("LIMIT [%d]", g.step_limits[0])
		}
	case 1:
		stepPart = fmt.Sprintf("%d STEPS", g.steps[0])
		if len(g.step_limits) > 0 {
			if len(g.step_limits) != g.steps[0] {
				return ngql, errors.New("SemanticError: length must be equal to GO step size")
			}

			ls := make([]string, 0, len(g.step_limits))
			for idx := range g.step_limits {
				ls = append(ls, strconv.Itoa(g.step_limits[idx]))
			}

			stepLimits = fmt.Sprintf("LIMIT [%s]", strings.Join(ls, ", "))
		}

	case 2:
		if g.steps[1] < g.steps[0] {
			return ngql, errors.New("SemanticError: upper bound steps should be greater than lower bound")
		}

		stepPart = fmt.Sprintf("%d TO %d STEPS", g.steps[0], g.steps[1])
		if len(g.step_limits) > 0 {
			if len(g.step_limits)-1 != g.steps[1]-g.steps[0] {
				return ngql, errors.New("SemanticError: length must be equal to GO step size")
			}

			ls := make([]string, 0, len(g.step_limits))
			for idx := range g.step_limits {
				ls = append(ls, strconv.Itoa(g.step_limits[idx]))
			}

			stepLimits = fmt.Sprintf("LIMIT [%s]", strings.Join(ls, ", "))
		}
	}

	if len(g.yields) > 0 {
		yieldPart = "YIELD " + strings.Join(g.yields, ", ")
	}

	if g.limit > 0 {
		if g.offset > 0 {
			pipeLimitPart = fmt.Sprintf(" | LIMIT %d OFFSET %d", g.limit, g.offset)
		} else {
			pipeLimitPart = fmt.Sprintf(" | LIMIT %d", g.limit)
		}
	}

	ngql = fmt.Sprintf("GO %s FROM %s OVER %s %s %s %s", stepPart, strings.Join(g.from, ", "), g.over, yieldPart, stepLimits, pipeLimitPart)

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
