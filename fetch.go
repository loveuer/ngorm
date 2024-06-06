package ngorm

import (
	"fmt"
	"reflect"
	"strings"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type fetchController struct {
	//client *Client
	sess  *session
	model any
	ngql  string
	key   string
	ids   []string
	tags  []string
}

func (f *fetchController) Model(model any) *fetchController {
	f.model = model
	return f
}

func (f *fetchController) Tags(tags ...string) *fetchController {
	f.tags = tags
	return f
}

func (f *fetchController) Key(key string) *fetchController {
	f.key = key
	return f
}

func (f *fetchController) ctorNGQL() error {
	if f.model != nil {
		if model, err := parse(f.model); err == nil {
			if model.rt.Kind() == reflect.Struct {
				ps := make([]string, 0, len(model.tags))
				for k := range model.tags {
					if k != "VertexID" {
						ps = append(ps, k)
					}
				}

				if !(len(f.tags) > 0 && len(f.tags) < len(ps)) {
					f.tags = ps
				}
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
	if err := f.ctorNGQL(); err != nil {
		return nil, err
	}

	return f.sess.client.Raw(f.ngql).RawResult()
}

func (f *fetchController) Result() (any, error) {
	if err := f.ctorNGQL(); err != nil {
		return nil, err
	}

	return f.sess.client.Raw(f.ngql).Result()
}

func (f *fetchController) Scan(dest any) error {
	if f.model == nil {
		f.model = dest
	}

	if err := f.ctorNGQL(); err != nil {
		return err
	}

	return f.sess.client.Raw(f.ngql).Scan(dest)
}

// Deprecated: use Scan instead
func (f *fetchController) Find(dest any) error {
	return f.Scan(dest)
}
