package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"strings"
)

type fetchController struct {
	client *Client

	ngql  string
	key   string
	ids   []string
	tags  []string
	props []string
}

func (f *fetchController) Props(props ...string) *fetchController {
	f.props = props
	return f
}

func (f *fetchController) Tags(tags ...string) *fetchController {
	f.tags = tags
	return f
}

func (f *fetchController) RawResult() (*nebula.ResultSet, error) {
	return f.client.Raw(f.ngql).RawResult()
}

func (f *fetchController) Result() (any, error) {
	return f.client.Raw(f.ngql).Result()
}

func (f *fetchController) Scan(dest any) error {
	return f.client.Raw(f.ngql).Scan(dest)
}

func (fc *fetchController) genngql(model any) (string, error) {
	var (
	//err error
	)

	if len(fc.ids) == 0 {
		return "", ErrSyntax("ids length must greater than 0")
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
