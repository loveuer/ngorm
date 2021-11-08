package ngorm

import (
	"errors"
	"fmt"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type Relationship struct {
	From string `json:"from"`
	To   string `json:"to"`
	Edge string `json:"edge"`
}

type PathController struct {
	db   *NGDB
	path string
	from string
	to   string
	over string
	upto int
}

func (db *NGDB) Path(pathKind string) *PathController {
	var (
		pe = &PathController{db: db, path: strings.ToUpper(pathKind)}
	)

	return pe
}

func (pe *PathController) From(from string) *PathController {
	pe.from = from

	return pe
}

func (pe *PathController) To(to string) *PathController {
	pe.to = to

	return pe
}

func (pe *PathController) Over(over string) *PathController {
	pe.over = over

	return pe
}

func (pe *PathController) Upto(num int) *PathController {
	pe.upto = num

	return pe
}

func (pe *PathController) genngql() (ngql string, err error) {
	if pe.from == "" {
		err = errors.New("from value can't be empty")
		return
	}

	if pe.to == "" {
		err = errors.New("to value can't be empty")
		return
	}

	if pe.over == "" {
		err = errors.New("over value can't be empty")
		return
	}

	switch pe.path {
	case "SHORTEST", "ALL", "NOLOOP":
		//	pass
	default:
		err = errors.New("path only accept 'SHORTEST | ALL | NOLOOP' 3 kinds")
		return
	}

	ngql = fmt.Sprintf("FIND %s PATH FROM '%s' TO '%s' OVER %s", pe.path, pe.from, pe.to, pe.over)
	if pe.upto > 0 {
		ngql = fmt.Sprintf("%s UPTO %d STEPS", ngql, pe.upto)
	}

	return
}

func (pe *PathController) Value() (*nebula.ResultSet, error) {
	e := &entry{db: pe.db, ctrl: pe}

	return e.value()
}

func (pe *PathController) Find() ([][]Relationship, error) {
	var result [][]Relationship

	e := &entry{db: pe.db, ctrl: pe}

	set, err := e.value()
	if err != nil {
		return result, err
	}

	for idx := 0; idx < set.GetRowSize(); idx++ {
		record, err := set.GetRowValuesByIndex(idx)
		if err != nil {
			return result, err
		}

		val, err := record.GetValueByIndex(0)
		if err != nil {
			return result, err
		}

		path, err := val.AsPath()
		if err != nil {
			return result, err
		}

		relations := path.GetRelationships()

		rs := make([]Relationship, 0)

		for _, r := range relations {
			src, err := r.GetSrcVertexID().AsString()
			if err != nil {
				break
			}
			dst, err := r.GetDstVertexID().AsString()
			if err != nil {
				break
			}
			rs = append(rs, Relationship{From: src, To: dst, Edge: r.GetEdgeName()})
		}

		result = append(result, rs)
	}

	return result, nil
}
