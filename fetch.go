package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"reflect"
	"strings"
)

type fetchController struct {
	client *Client

	ngql string
	key  string
	ids  []string
	tags []string
}

func (f *fetchController) Tags(tags ...string) *fetchController {
	f.tags = tags
	return f
}

func (f *fetchController) Key(key string) *fetchController {
	f.key = key
	return f
}

func (f *fetchController) ctorNGQL(models ...any) error {
	if len(models) > 0 && models[0] != nil {
		if model, err := parse(models[0]); err == nil {
			if model.rt.Kind() == reflect.Struct {
				ps := make([]string, 0, len(model.tags))
				for k := range model.tags {
					ps = append(ps, k)
				}
				f.tags = ps
			}
		}
	}

	if f.key == "" {
		f.key = "v"
	}

	fields := make([]string, 0, len(f.tags))
	vids := make([]string, 0, len(f.ids))
	for _, tag := range f.tags {
		fields = append(fields, fmt.Sprintf("%s.%s AS %s", tag, f.key, tag))
	}
	for _, id := range f.ids {
		vids = append(vids, fmt.Sprintf("'%s'", id))
	}

	f.ngql = fmt.Sprintf("FETCH PROP ON %s %s YIELD id(vertex) as VertexID, %s",
		strings.Join(f.tags, ", "),
		strings.Join(vids, ", "),
		strings.Join(fields, ", "),
	)

	return nil
}

func (f *fetchController) RawResult() (*nebula.ResultSet, error) {
	return f.client.Raw(f.ngql).RawResult()
}

func (f *fetchController) Result() (any, error) {
	return f.client.Raw(f.ngql).Result()
}

func (f *fetchController) Scan(dest any) error {
	if err := f.ctorNGQL(dest); err != nil {
		return err
	}

	return f.client.Raw(f.ngql).Scan(dest)
}

func (fc *fetchController) genngql(model any) (string, error) {
	var (
	//err error
	)

	if len(fc.ids) == 0 {
		return "", fmt.Errorf("%w: ids length must greater than 0", ErrSyntax)
	}

	ids := strings.Join(fc.ids, ", ")

	if fc.key == "" {
		fc.key = "v"
	}

	if len(fc.tags) == 1 && fc.tags[0] == "*" {
		return fmt.Sprintf("FETCH PROP ON * %s YIELD vertex AS v", ids), nil
	}

	if len(fc.tags) == 0 {
		//if fc.tags, err = fc.getTags(model); err != nil {
		//	return "", err
		//}
		//
		//logrus.Debugf("compatible get tags: %v", fc.tags)
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
