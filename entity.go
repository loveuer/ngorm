package ngorm

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"strings"
)

type entity struct {
	c         *Client
	logger    logger
	ngql      string
	set       *nebula.ResultSet
	err       error
	formatted []map[string]any
}

func (e *entity) execute(retry int) {

	e.logger.Debug(fmt.Sprintf("[ngorm] [sessions: %d] execute '%s'", e.c.client.GetTotalSessionCount(), e.ngql))

	e.set, e.err = e.c.client.Execute(e.ngql)

	if e.err != nil {
		e.logger.Debug(fmt.Sprintf("[ngorm] execute '%s' err: %v", e.ngql, e.err))

		if strings.Contains(e.err.Error(), "EOF") && retry == 0 {
			e.logger.Debug("[ngorm] EOF: reconnection...")

			var (
				err error
			)

			if e.c.client, e.err = nebula.NewSessionPool(*config, cc.Logger); err != nil {
				e.logger.Debug(fmt.Sprintf("[ngorm] renew session pool err: %v", e.err))
				return
			}

			e.execute(retry + 1)
		}
	}
}

func (e *entity) formatSet() {
	var (
		columns   = e.set.GetColNames()
		formatted = make([]map[string]any, e.set.GetRowSize())
	)

	for rowIndex := 0; rowIndex < e.set.GetRowSize(); rowIndex++ {
		var record *nebula.Record

		if record, e.err = e.set.GetRowValuesByIndex(rowIndex); e.err != nil {
			e.logger.Debug(fmt.Sprintf("[ngorm] get row: %d value err: %v", rowIndex, e.err))
			return
		}

		formatted[rowIndex] = make(map[string]any, len(columns))

		for _, column := range columns {
			var (
				vw   *nebula.ValueWrapper
				data any
			)
			if vw, e.err = record.GetValueByColName(column); e.err != nil {
				e.logger.Debug(fmt.Sprintf("[ngorm] get row: %d column: %s err: %v", rowIndex, column, e.err))
				return
			}

			if data, e.err = e.handleValueWrapper(vw); e.err != nil {
				return
			}

			formatted[rowIndex][column] = data
		}
	}

	e.formatted = formatted
}

func (e *entity) RawResult() (*nebula.ResultSet, error) {
	e.execute(0)
	return e.set, e.err
}

func (e *entity) Result() (any, error) {
	e.execute(0)
	if e.err != nil {
		return nil, e.err
	}

	e.formatSet()

	if len(e.formatted) == 1 {
		return e.formatted[0], e.err
	}

	return e.formatted, e.err
}
