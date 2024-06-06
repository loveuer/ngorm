package ngorm

import (
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"strings"
	"time"
)

type entity struct {
	sess      *Session
	ngql      string
	set       *nebula.ResultSet
	err       error
	formatted []map[string]any
}

func (e *entity) execute() {

	clog.l.Debug("ngql: '%s'", e.ngql)

	retry := 0
EXECUTE:
	e.set, e.err = e.sess.client.pool.Execute(e.ngql)

	if e.err != nil {
		clog.l.Debug("execute ngql: '%s', err: %v", e.ngql, e.err)
		if strings.Contains(e.err.Error(), "EOF") || strings.Contains(e.err.Error(), "broken pipe") {
		RETRY:
			if retry < e.sess.cfg.MaxRetry {
				retry++
				clog.l.Warn("nebula connection EOF(broken pipe): reconnect after %d seconds...[%d/%d]", 1<<retry, retry, e.sess.cfg.MaxRetry)

				time.Sleep(time.Duration(1<<retry) * time.Second)

				if e.sess.client.pool, e.err = nebula.NewSessionPool(*config, clog); e.err != nil {
					e.sess.logger.Debug("reconnect err: %v", e.err)
					goto RETRY
				}

				goto EXECUTE
			}
			// if max_retry = 0(no retry), end
		}
		// if err not EOF(broken pipe...), no retry, end
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
			clog.l.Debug("get row: %d value err: %v", rowIndex, e.err)
			return
		}

		formatted[rowIndex] = make(map[string]any, len(columns))

		for _, column := range columns {
			var (
				vw   *nebula.ValueWrapper
				data any
			)
			if vw, e.err = record.GetValueByColName(column); e.err != nil {
				clog.l.Debug("get row: %d column: %s err: %v", rowIndex, column, e.err)
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
	e.execute()
	return e.set, e.err
}

func (e *entity) Result() (any, error) {
	e.execute()
	if e.err != nil {
		return nil, e.err
	}

	e.formatSet()

	if len(e.formatted) == 1 {
		return e.formatted[0], e.err
	}

	return e.formatted, e.err
}
