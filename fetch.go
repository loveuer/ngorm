package ngorm

import (
	"errors"
	"fmt"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type FetchController struct {
	db      *NGDB
	session *nebula.Session
	tags    []string
	ids     []string
	key     string
	ngql    string
	err     error
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

func (fc *FetchController) genngql() (string, error) {
	if len(fc.tags) == 0 {
		return "", errors.New("tags length must greater than 0")
	}

	t := strings.Join(fc.tags, ", ")

	if len(fc.ids) == 0 {
		return "", errors.New("ids lenght must greater than 0")
	}

	ids := strings.Join(fc.ids, ", ")

	if fc.key == "" {
		return "", errors.New("empty tag key")
	}

	fields := make([]string, 0, len(fc.tags))
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
	return e.find(model)
}
