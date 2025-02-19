package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"reflect"
	"strings"
)

type _path struct {
	src, dst string
	rank     int
}

type fetchPathController struct {
	//client *Client
	sess  *Session
	model any
	edge  string
	ngql  string
	yield []string
	tags  []string
	paths []*_path
}

func (f *fetchPathController) Model(model any) *fetchPathController {
	f.model = model
	return f
}

func (f *fetchPathController) Edge(edge string) *fetchPathController {
	f.edge = edge
	return f
}

func (f *fetchPathController) Path(src string, dst string, ranks ...int) *fetchPathController {
	rank := 0
	if len(ranks) > 0 && ranks[0] > 0 {
		rank = ranks[0]
	}
	f.paths = append(f.paths, &_path{src, dst, rank})
	return f
}

func (f *fetchPathController) Yield(yield string, as ...string) *fetchPathController {
	item := yield
	if len(as) > 0 && len(as[0]) > 0 {
		item = fmt.Sprintf("%s AS %s", yield, as[0])
	}

	f.yield = append(f.yield, item)

	return f
}

func (f *fetchPathController) ctorNGQL() error {
	if f.model != nil {
		if model, err := parse(f.model); err == nil {
			if model.rt.Kind() == reflect.Struct {
				ps := make([]string, 0, len(model.tags))
				for k := range model.tags {
					if k != "VertexID" {
						ps = append(ps, k)
					}
				}
			}
		}
	}

	paths := make([]string, 0, len(f.paths))
	for _, path := range f.paths {
		if path.src == "" || path.dst == "" {
			return fmt.Errorf("invalid path: src or dst required")
		}

		var item string
		if path.rank == 0 {
			item = fmt.Sprintf("'%s' -> '%s'", path.src, path.dst)
		} else if path.rank > 0 {
			item = fmt.Sprintf("'%s' -> '%s'@%d", path.src, path.dst, path.rank)
		} else {
			return fmt.Errorf("invalid path: rank invalid")
		}

		paths = append(paths, item)
	}

	if len(paths) == 0 {
		return fmt.Errorf("at least one path required")
	}

	if len(f.yield) == 0 {
		return fmt.Errorf("at least one yield required")
	}

	f.ngql = fmt.Sprintf("FETCH PROP ON %s %s YIELD %s",
		f.edge,
		strings.Join(paths, ", "),
		strings.Join(f.yield, ", "),
	)

	return nil
}

func (f *fetchPathController) RawResult() (*nebula.ResultSet, error) {
	if err := f.ctorNGQL(); err != nil {
		return nil, err
	}

	return f.sess.client.Raw(f.ngql).RawResult()
}

func (f *fetchPathController) Result() (any, error) {
	if err := f.ctorNGQL(); err != nil {
		return nil, err
	}

	return f.sess.client.Raw(f.ngql).Result()
}

func (f *fetchPathController) Scan(dest any) error {
	if f.model == nil {
		f.model = dest
	}

	if err := f.ctorNGQL(); err != nil {
		return err
	}

	return f.sess.client.Raw(f.ngql).Scan(dest)
}

// Deprecated: use Scan instead
func (f *fetchPathController) Find(dest any) error {
	return f.Scan(dest)
}
