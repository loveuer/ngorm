package ngorm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type direction int

const (
	TwoDirection     direction = 0 //双向
	ForwardDirection direction = 1 //正向
	ReverseDirection direction = 2 //反向
)

var (
	directionMap = map[direction]string{
		TwoDirection:     "--",
		ForwardDirection: "-->",
		ReverseDirection: "<--",
	}
)

type drop struct {
	Name  string
	Dire  direction
	Value any
}
type factor struct {
	Query string
	Value []interface{}
}
type Limit struct {
	Limit  *int
	Offset int
}
type matchPathController struct {
	matchController
}

type matchController struct {
	//client  *Client
	sess    *session
	points  []drop
	ngql    string
	where   string
	rsql    string
	factor  []factor
	selects []string
	limit   Limit
	order   []string
	key     string
	edgs    map[string]struct{}
}

func (m *matchController) With(value any, name string, dire direction) *matchController {
	m.points = append(m.points, drop{
		Dire:  dire,
		Value: value,
		Name:  name,
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
func (m *matchController) Key(key string) *matchController {
	m.key = key
	return m
}
func (m *matchController) Where(query string, args ...interface{}) *matchController {
	m.factor = append(m.factor, factor{
		Query: query,
		Value: args,
	})
	return m
}
func (m *matchController) Select(value ...string) *matchController {
	m.selects = append(m.selects, value...)
	return m
}
func (m *matchController) Finds(value ...any) error {
	if len(m.selects) == 0 {
		return ErrSelectsInvalid
	}
	m.ctorNGQL()
	m.mReturnSql("finds")
	m.ngql += fmt.Sprintf("return %v", strings.TrimSuffix(m.rsql, ","))
	m.mOrder()
	m.mLimit()
	//fmt.Println("sql:", m.ngql)
	return m.sess.client.Raw(m.ngql).Finds(m.key, value...)
}

func (m *matchController) Count(value ...*int64) error {
	if len(m.selects) == 0 {
		return ErrSelectsInvalid
	}
	m.ctorNGQL()
	m.mReturnSql("count")
	m.ngql += fmt.Sprintf("return %v", strings.TrimSuffix(m.rsql, ","))
	//fmt.Println("sql:", m.ngql)
	return m.sess.client.Raw(m.ngql).Count(value...)
}

func (m *matchController) mReturnSql(rType string) {
	var (
		formatMap = make(map[string]string)
	)
	switch rType {
	case "count":
		formatMap["point"] = "count(%v),"
		formatMap["edge"] = "count(e%v),"
	case "finds":
		formatMap["point"] = "%v,"
		formatMap["edge"] = "e%v,"
	}
	for i := range m.selects {
		for pi := range m.points {
			if pi > 0 {
				if m.selects[i] == m.points[pi-1].Name+m.points[pi].Name {
					m.edgs[m.points[pi].Name] = struct{}{}
					m.rsql += fmt.Sprintf(formatMap["edge"], m.points[pi].Name)
					break
				}
			}
			if m.selects[i] == m.points[pi].Name {
				m.rsql += fmt.Sprintf(formatMap["point"], m.selects[i])
				break
			} else if strings.Contains(m.selects[i], m.points[pi].Name+".") {
				m.rsql += fmt.Sprintf(formatMap["point"], m.selects[i]+"."+m.key)
			}
		}
	}
}

func (m *matchController) ctorNGQL() {
	for i := range m.selects {
		for pi := range m.points {
			if pi == 0 {
				continue
			}
			if m.selects[i] == m.points[pi-1].Name+m.points[pi].Name {
				m.edgs[m.points[pi].Name] = struct{}{}
				break
			}
		}
	}
	m.ngql = fmt.Sprintf("match (%v) ", m.points[0].Name)
	for _, point := range m.points[1:] {
		format := directionMap[point.Dire] + "(%v)"
		if _, ok := m.edgs[point.Name]; ok {
			format = strings.ReplaceAll(format, "--", "-[e%v]-")
			m.ngql += fmt.Sprintf(format, point.Name, point.Name)
		} else {
			m.ngql += fmt.Sprintf(format, point.Name)
		}
	}
	for i := range m.factor {
		if i == 0 {
			m.where = fmt.Sprintf(" (%v)", parseFactor(m.factor[i], m.key))
		} else {
			m.where = fmt.Sprintf("and (%v)", parseFactor(m.factor[i], m.key))
		}
	}
	m.ngql += fmt.Sprintf(" where %v", m.where)
}
func repQuery(query, key string) string {
	reg := regexp.MustCompile(`([0-9A-Za-z]+\.[0-9A-Za-z]+(\.[0-9A-Za-z]+)?)`)
	colRowslice := reg.FindStringSubmatch(query)
	if len(colRowslice) == 3 {
		if colRowslice[2] == "" {
			return strings.ReplaceAll(query, colRowslice[2], fmt.Sprintf("%v.%v", colRowslice[2], key))
		}
	}
	return query
}
func parseFactor(factor factor, key string) string {
	var (
		nsql      string
		queryList []string
	)
	if len(factor.Value) == 0 {
		return factor.Query
	}
	if strings.Contains(factor.Query, "?") {
		queryList = strings.Split(factor.Query, "?")
		for i := range queryList {
			head := reflect.ValueOf(factor.Value[i])
			switch head.Type().Elem().Kind() {
			case reflect.Slice, reflect.Array:
				valueByte, _ := json.Marshal(factor.Value[i])
				nsql += fmt.Sprintf("%v %v ", repQuery(queryList[i], key), string(valueByte))
			case reflect.String:
				nsql += fmt.Sprintf("%v '%v' ", repQuery(queryList[i], key), factor.Value[i])
			case reflect.Int8, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
				nsql += fmt.Sprintf("%v %v ", repQuery(queryList[i], key), factor.Value[i])
			}
		}
	} else {
		queryList = strings.Split(factor.Query, ",")
		for i := range queryList {
			head := reflect.ValueOf(factor.Value[i])
			sql := factorAssemble(queryList[i], key, head)
			if i > 0 {
				sql = " and " + sql
			}
			nsql += sql
		}
	}
	return nsql
}

func factorAssemble(factorQ, key string, value reflect.Value) string {
	addkey := func(factorQ, key string) string {
		if strings.Contains(factorQ, "id") {
			return factorQ
		}
		if !strings.HasSuffix(factorQ, key) {
			return factorQ + "." + key
		}
		return factorQ
	}
	factorQ = strings.ReplaceAll(factorQ, " ", "")
	head := value.Type()
	switch head.Kind() {
	case reflect.Slice, reflect.Array:
		valueByte, _ := json.Marshal(value.Interface())
		return fmt.Sprintf("%v in %v ", addkey(factorQ, key), string(valueByte))
	case reflect.String:
		return fmt.Sprintf("%v == '%v' ", addkey(factorQ, key), head)
	case reflect.Int8, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
		return fmt.Sprintf("%v == %v ", addkey(factorQ, key), head)
	case reflect.Map:
		newValue := value.Interface()
		for k, v := range newValue.(map[string]interface{}) {
			return factorAssemble(fmt.Sprintf("%v.%v", factorQ, k), key, reflect.ValueOf(v))
		}
	case reflect.Struct:
		for idx := 0; idx < head.NumField(); idx++ {
			tag := strings.TrimSpace(head.Field(idx).Tag.Get("nebula"))
			if tag == "" || tag == "-" {
				continue
			}
			return factorAssemble(fmt.Sprintf("%v.%v", factorQ, tag), key, value.FieldByName(head.Field(idx).Name))
		}
	}
	return ""
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
