package ngorm

import "fmt"

func (c *Client) Raw(ngql string) *entity {
	var (
		e = &entity{client: c}
	)

	e.set, e.err = c.client.Execute(ngql)
	if e.err != nil {
		c.logger.Debug(fmt.Sprintf("[ngorm] execute '%s' err: %v", ngql, e.err))
	}

	return e
}
