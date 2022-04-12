package ngorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
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
	fc.tags = tags
	return fc
}

// Key Tag properties.k-v key
func (fc *FetchController) Key(key string) *FetchController {
	fc.key = key
	return fc
}

func (fc *FetchController) genngql(model interface{}) (string, error) {
	if len(fc.ids) == 0 {
		return "", ErrorSyntaxGen("ids length must greater than 0")
	}

	ids := strings.Join(fc.ids, ", ")

	if fc.key == "" {
		return "", ErrorSyntaxGen("fetch tags must specify key")
	}

	if len(fc.tags) == 1 && fc.tags[0] == "*" {
		return fmt.Sprintf("FETCH PROP ON * %s YIELD vertex AS v", ids), nil
	}

	if len(fc.tags) == 0 {
		if fc.tags, err = fc.getTags(model); err != nil {
			return "", err
		}

		log.Debugf("compatible get tags: %v", fc.tags)
	}

	t := strings.Join(fc.tags, ", ")

	var (
		fields = make([]string, 0, len(fc.tags))
	)

	for _, field := range fc.tags {
		fields = append(fields, fmt.Sprintf("%s.%s as %s", field, fc.key, field))
	}

	f := strings.Join(fields, ", ")
	return fmt.Sprintf("fetch PROP on %s %s yield id(vertex) as VertexID, %s", t, ids, f), nil
}

func (fc *FetchController) getTags(model interface{}) ([]string, error) {
	var (
		ok      bool
		tags    = make([]string, 0)
		rt      reflect.Type
		exclude = map[string]struct{}{
			"VertexID": {},
		}
	)

	rv := reflect.ValueOf(model)

	if rv.Type().Kind() != reflect.Ptr {
		log.Errorf("model type not ptr, but: %s", rv.Type().String())
		return tags, ErrorModelNotPtr
	}

	rv = rv.Elem()

	if !rv.IsValid() {
		return tags, ErrorInvalidModel
	}

	if rv.Type().Kind() == reflect.Slice || rv.Type().Kind() == reflect.Array {
		rt = rv.Type().Elem()
	} else {
		rt = rv.Type()
	}

	for index := 0; index < rt.NumField(); index++ {
		ftag := rt.Field(index).Tag.Get("nebula")
		fmt.Printf("tag: %s\n", ftag)
		if tag := strings.TrimSpace(ftag); tag != "" && tag != "-" {
			if _, ok = exclude[tag]; !ok {
				tags = append(tags, tag)
			}
		}
	}

	return tags, nil
}

func (fc *FetchController) Find(model interface{}) error {
	var (
		err error
		e   = &entry{db: fc.db}
	)

	if fc.ngql, err = fc.genngql(model); err != nil {
		return err
	}

	if fc.ngql == "" {
		// todo ErrorType
		return errors.New("empty ngql")
	}

	e.ngql = fc.ngql

	if len(fc.tags) == 1 && fc.tags[0] == "*" {
		return e.finds(model)
	}

	return e.find(model)
}

type FetchPathController struct {
	db    *NGDB
	edge  string
	paths []string
	key   string
	ngql  string
	err   error
}

func (db *NGDB) FetchPath(edge string) *FetchPathController {
	return &FetchPathController{
		db:   db,
		edge: edge,
	}
}

func (fp *FetchPathController) Key(key string) *FetchPathController {
	fp.key = key
	return fp
}

func (fp *FetchPathController) Path(paths ...string) *FetchPathController {
	fp.paths = paths
	return fp
}

func (fp *FetchPathController) genngql() (string, error) {
	if fp.edge == "" {
		return "", ErrorSyntaxGen("edge can't be ''")
	}

	if len(fp.paths) == 0 {
		return "", ErrorSyntaxGen("length of paths can't be 0")
	}

	if fp.key == "" {
		return "", ErrorSyntaxGen("key can't be ''")
	}

	p := strings.Join(fp.paths, ", ")

	return fmt.Sprintf("FETCH PROP ON %s %s YIELD %s.%s as v", fp.edge, p, fp.edge, fp.key), nil
}

func (fp *FetchPathController) Find(model interface{}) error {
	e := &entry{db: fp.db, ctrl: fp}

	return e.find(model)
}
