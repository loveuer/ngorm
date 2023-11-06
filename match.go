package ngorm

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type direction int

const (
	TwoDirection     direction = 0 //双向
	ForwardDirection direction = 1 //正向
	ReverseDirection direction = 2 //反向
)

type drop struct {
	Number string
	Dire   direction
	Value  any
}
type Where struct {
	Query string
	Value []interface{}
}
type Limit struct {
	Limit  *int
	Offset int
}
type matchController struct {
	client *Client
	head   drop
	models []drop
	ngql   string
	where  string
	limit  Limit
	order  []string
}

func (m *matchController) With(dire direction, value any) *matchController {
	m.models = append(m.models, drop{
		Number: "v2",
		Dire:   dire,
		Value:  value,
	})
	return m
}

func (m *matchController) Limit(value int) *matchController {
	m.limit.Limit = new(int)
	*(m.limit.Limit) = value
	return m
}

func (m *matchController) Offset(value int) *matchController {
	if m.limit.Limit == nil {
		m.limit.Limit = new(int)
		*m.limit.Limit = 1000
	}
	m.limit.Offset = value
	return m
}
func (m *matchController) Order(value string) *matchController {
	m.order = append(m.order, value)
	return m
}

func (m *matchController) CountPath(value *int64) error {
	m.ctorNGQL()
	m.ngql += fmt.Sprintf(" return count(e%s)", m.models[0].Number)
	fmt.Println("sql:", m.ngql)
	m.client.Raw(m.ngql).Count(value)
	return nil
}
func (m *matchController) FindPath(value any) error {
	m.ctorNGQL()
	m.ngql += fmt.Sprintf(" return e%s", m.models[0].Number)
	m.mOrder()
	m.mLimit()
	fmt.Println("sql:", m.ngql)
	m.client.Raw(m.ngql).FindPath(fmt.Sprintf("e%s", m.models[0].Number), value)
	return nil
}
func (m *matchController) ctorNGQL() {
	m.ngql = "match (head)"
	for i := range m.models {
		switch m.models[i].Dire {
		case TwoDirection:
			m.ngql += fmt.Sprintf("-[e%v]-(%v)", m.models[i].Number, m.models[i].Number)
		case ForwardDirection:
			m.ngql += fmt.Sprintf("-[e%v]->(%v)", m.models[i].Number, m.models[i].Number)
		case ReverseDirection:
			m.ngql += fmt.Sprintf("<-[e%v]-(%v)", m.models[i].Number, m.models[i].Number)
		}

	}

	if head := parseWhere(m.head); head != "" {
		m.where = fmt.Sprintf(" where %v ", parseWhere(m.head))
	}
	{
		for _, model := range m.models {
			if where := parseWhere(model); where != "" {
				if m.where == "" {
					m.where = fmt.Sprintf(" where %v ", where)
				} else {
					m.where += fmt.Sprintf(" and %v ", where)
				}
			}

		}
	}
	m.ngql += m.where

}

func parseWhere(model drop) string {
	var (
		nsql string
	)
	fmt.Println(model.Value)
	head := reflect.ValueOf(model.Value)
	switch head.Type().Elem().Kind() {
	case reflect.Slice:
		fmt.Println("model.Value:", model.Value)
		headByte, _ := json.Marshal(model.Value)
		if string(headByte) != "[]" {
			nsql = fmt.Sprintf("id(%v) in %v", model.Number, string(headByte))
		}
	}
	return nsql
}

func (m *matchController) mOrder() {
	if len(m.order) > 0 {
		m.ngql += " ORDER BY "
		for i, v := range m.order {
			if i == len(m.order)-1 {
				m.ngql += fmt.Sprintf("%v ", v)
			} else {
				m.ngql += fmt.Sprintf("%v,", v)
			}

		}
	}
}

func (m *matchController) mLimit() {
	if m.limit.Limit != nil {
		limit := ""
		if m.limit.Offset != 0 {
			limit += fmt.Sprintf(" SKIP %v", m.limit.Offset)
		}
		limit += fmt.Sprintf(" LIMIT %v ", *m.limit.Limit)
		m.ngql += limit
	}
}
