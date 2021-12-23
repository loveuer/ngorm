package ngorm

import (
	"errors"
	"fmt"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type FetchController struct {
	db   *NGDB
	tags []string
	ids  []string
	key  string
	ngql string
	err  error
}

func (db *NGDB) Fetch(ids ...string) *FetchController {
	fc := &FetchController{
		db: db,
	}

	for _, id := range ids {
		fc.ids = append(fc.ids, fmt.Sprintf("'%s'", id))
	}

	return fc
}

func (fc *FetchController) Tags(tags ...string) *FetchController {
	tmp := make([]string, 0)
	for idx := range tags {
		if strings.ToLower(tags[idx]) != "message_count" {
			tmp = append(tmp, tags[idx])
		}
	}

	fc.tags = tmp
	return fc
}

// Key Tag properties.k-v key
func (fc *FetchController) Key(key string) *FetchController {
	fc.key = key
	return fc
}

func (fc *FetchController) genngql() (string, error) {
	if len(fc.ids) == 0 {
		return "", errors.New("ids length must greater than 0")
	}

	ids := strings.Join(fc.ids, ", ")

	if fc.key == "" {
		return "", errors.New("fetch tags must specify key")
	}

	if len(fc.tags) == 0 {
		return fmt.Sprintf("FETCH PROP ON * %s", ids), nil
	}

	t := strings.Join(fc.tags, ", ")

	var (
		fields = make([]string, 0, len(fc.tags))
	)

	for _, field := range fc.tags {
		fields = append(fields, fmt.Sprintf("%s.%s as %s", field, fc.key, field))
	}

	f := strings.Join(fields, ", ")
	return fmt.Sprintf("fetch PROP on %s %s yield %s", t, ids, f), nil
}

func (f *FetchController) Value() (*nebula.ResultSet, error) {
	e := &entry{db: f.db, ctrl: f}
	return e.value()
}

func (f *FetchController) Find(model interface{}) error {
	e := &entry{db: f.db, ctrl: f}

	if len(f.tags) == 0 {
		return e.finds(model)
	}

	return e.find(model)
}

type FetchPathController struct {
	db    *NGDB
	edge  string
	props []string
	paths []string
	ngql  string
	err   error
}

func (db *NGDB) FetchPath(edge string) *FetchPathController {
	return &FetchPathController{
		db:   db,
		edge: edge,
	}
}

func (fp *FetchPathController) Props(props ...string) *FetchPathController {
	fp.props = props
	return fp
}

func (fp *FetchPathController) Path(paths ...string) *FetchPathController {
	fp.paths = paths
	return fp
}

func (fp *FetchPathController) genngql() (string, error) {
	if fp.edge == "" {
		return "", errors.New("edge can't be ''")
	}

	if len(fp.paths) == 0 {
		return "", errors.New("length of paths can't be 0")
	}

	if len(fp.props) == 0 {
		return "", errors.New("length of props can't be 0")
	}

	if fp.props[0] == "" {
		return "", errors.New("props can't be ''")
	}

	p := strings.Join(fp.paths, ", ")

	for idx := range fp.props {
		fp.props[idx] = fmt.Sprintf("%s.%s as %s", fp.edge, fp.props[idx], fp.props[idx])
	}

	y := strings.Join(fp.props, ", ")

	return fmt.Sprintf("FETCH PROP ON %s %s YIELD %s", fp.edge, p, y), nil
}

func (fp *FetchPathController) Find(model interface{}) error {
	e := &entry{db: fp.db, ctrl: fp}

	return e.find(model)
}
