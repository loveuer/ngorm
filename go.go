package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"reflect"
	"strings"
)

type EdgeType uint

const (
	EdgeTypeForward EdgeType = iota
	EdgeTypeReverse
	EdgeTypeBoth
)

type goController struct {
	//client   *Client
	sess     *Session
	model    any
	from     string
	steps    int
	edge     string
	edgeType string
	tags     []string
	key      string

	offset int
	limit  int

	ngql string
}

func (g *goController) Model(model any) *goController {
	g.model = model
	return g
}

func (g *goController) Steps(steps int) *goController {
	g.steps = steps
	return g
}

func (g *goController) Over(edge string, edgeType ...EdgeType) *goController {
	g.edge = edge

	if len(edgeType) > 0 {
		switch edgeType[0] {
		case EdgeTypeReverse:
			g.edgeType = " REVERSELY"
		case EdgeTypeBoth:
			g.edgeType = " BIDIRECT"
		}

	}

	return g
}

func (g *goController) Key(key string) *goController {
	g.key = key
	return g
}

func (g *goController) Tags(tags ...string) *goController {
	g.tags = tags
	return g
}

func (g *goController) Offset(offset int) *goController {
	g.offset = offset
	return g
}

func (g *goController) Limit(limit int) *goController {
	g.limit = limit
	return g
}

func (g *goController) ctorNGQL() error {
	if g.model != nil {
		if model, err := parse(g.model); err == nil {
			if model.rt.Kind() == reflect.Struct {
				ps := make([]string, 0, len(model.tags))
				for k := range model.tags {
					if k != "VertexID" {
						ps = append(ps, k)
					}
				}

				if !(len(g.tags) > 0 && len(g.tags) < len(ps)) {
					g.tags = ps
				}
			}
		}
	}

	if g.steps == 0 {
		g.steps = 1
	}

	if g.key == "" {
		g.key = "v"
	}

	fs := make([]string, 0, len(g.tags))

	for _, t := range g.tags {
		f := fmt.Sprintf("$$.%s.%s AS %s", t, g.key, t)
		fs = append(fs, f)
	}

	g.ngql = fmt.Sprintf("GO %d STEPS FROM '%s' OVER %s%s YIELD %s._dst as VertexID, %s ",
		g.steps,
		g.from,
		g.edge,
		g.edgeType,
		g.edge,
		strings.Join(fs, ", "),
	)

	if g.limit > 0 || g.offset > 0 {
		g.ngql = fmt.Sprintf("%s | limit %d, %d", g.ngql, g.offset, g.limit)
	}

	return nil
}

func (g *goController) RawResult() (*nebula.ResultSet, error) {
	if err := g.ctorNGQL(); err != nil {
		return nil, err
	}

	return g.sess.client.Raw(g.ngql).RawResult()
}

func (g *goController) Result() (any, error) {
	if err := g.ctorNGQL(); err != nil {
		return nil, err
	}

	return g.sess.client.Raw(g.ngql).Result()
}

func (g *goController) Scan(dest any) error {
	if g.model == nil {
		g.model = dest
	}

	if err := g.ctorNGQL(); err != nil {
		return err
	}

	return g.sess.client.Raw(g.ngql).Scan(dest)
}

// Deprecated: use Scan instead
func (g *goController) Find(dest any) error {
	return g.Scan(dest)
}
