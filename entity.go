package ngorm

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"time"
)

type entity struct {
	client    *Client
	ngql      string
	set       *nebula.ResultSet
	err       error
	formatted any
}

func (e *entity) Scan(dest any) error {
	e.execute()
	if e.err != nil {
		return e.err
	}

	e.formatSet()
	if e.err != nil {
		return e.err
	}

	// todo scan
	model, err := Parse(dest)
	if err != nil {
		logrus.Debug(fmt.Sprintf("[ngorm] parse model err: %v", err))
		return err
	}

	return nil
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
	return e.formatted, e.err
}

func (e *entity) execute() {
	e.set, e.err = e.client.client.Execute(e.ngql)
	if e.err != nil {
		e.client.logger.Debug(fmt.Sprintf("[ngorm] execute '%s' err: %v", e.ngql, e.err))
	}
}

func (e *entity) formatSet() {
	var (
		columns   = e.set.GetColNames()
		formatted = make([]any, 0, len(columns))
	)

	for _, column := range columns {
		var (
			vws []*nebula.ValueWrapper
		)

		if vws, e.err = e.set.GetValuesByColName(column); e.err != nil {
			e.client.logger.Debug(fmt.Sprintf("[ngorm] get values by column name '%s' err: %v", column, e.err))
			continue
		}

		datas := make([]any, 0, len(vws))

		for _, vw := range vws {
			var data any
			if data, e.err = e.handleValueWrapper(vw); e.err != nil {
				return
			}

			datas = append(datas, data)
		}

		if len(datas) == 0 {
			continue
		}

		if len(datas) == 1 {
			formatted = append(formatted, datas[0])
		} else {
			formatted = append(formatted, datas)
		}
	}

	if len(formatted) == 1 {
		e.formatted = formatted[0]
	} else {
		e.formatted = formatted
	}
}

func (e *entity) handleValueWrapper(vw *nebula.ValueWrapper) (any, error) {
	switch {
	case vw.IsBool():
		val, _ := vw.AsBool()
		return val, nil
	case vw.IsFloat():
		val, _ := vw.AsFloat()
		return val, nil
	case vw.IsString():
		val, _ := vw.AsString()
		var parsed any
		if err := json.Unmarshal([]byte(val), &parsed); err == nil {
			return parsed, nil
		}
		return val, nil
	case vw.IsInt():
		val, _ := vw.AsInt()
		return val, nil
	case vw.IsDate():
		val, _ := vw.AsDate()
		t := time.Date(
			int(val.GetYear()),
			time.Month(val.GetMonth()),
			int(val.GetDay()),
			0, 0, 0, 0,
			time.Local,
		)
		return t, nil
	case vw.IsDateTime():
		var (
			val, _ = vw.AsDateTime()
		)

		d, err := val.GetLocalDateTimeWithTimezoneName(time.Local.String())
		if err != nil {
			e.client.logger.Debug(fmt.Sprintf("[ngorm] datetime value get local datetime err: %v", err))
			return nil, err
		}

		t := time.Date(
			int(d.GetYear()),
			time.Month(d.GetMonth()),
			int(d.GetDay()),
			int(d.GetHour()),
			int(d.GetMinute()),
			int(d.GetSec()),
			int(d.GetMicrosec()*1000),
			time.Local,
		)

		return t, nil
	case vw.IsDuration():
		val, _ := vw.AsDuration()
		d := time.Duration(val.GetSeconds()) * time.Second
		return d, nil
	case vw.IsEdge():
		// todo
		val, _ := vw.AsRelationship()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsEdge]")
	case vw.IsEmpty():
		return nil, nil
	case vw.IsGeography():
		// todo
		val, _ := vw.AsGeography()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsGeography]")
	case vw.IsList():
		val, _ := vw.AsList()
		list := make([]any, 0, len(val))

		for _, v := range val {
			data, err := e.handleValueWrapper(&v)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		return list, nil
	case vw.IsMap():
		val, _ := vw.AsMap()
		m := make(map[string]any, len(val))

		for k := range val {
			d := val[k]
			data, err := e.handleValueWrapper(&d)
			if err != nil {
				return nil, err
			}

			m[k] = data
		}

		return m, nil
	case vw.IsNull():
		return nil, nil
	case vw.IsPath():
		// todo
		val, _ := vw.AsPath()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsPath]")
	case vw.IsSet():
		val, _ := vw.AsDedupList()
		list := make([]any, 0, len(val))

		for _, v := range val {
			data, err := e.handleValueWrapper(&v)
			if err != nil {
				return nil, err
			}
			list = append(list, data)
		}

		return list, nil
	case vw.IsTime():
		// todo
		val, _ := vw.AsTime()
		_ = val
		logrus.Panic("impl this: handle nebula ValueWrapper[IsTime]")
	case vw.IsVertex():
		val, _ := vw.AsNode()
		tags := val.GetTags()
		m := make(map[string]map[string]any, len(tags))
		for _, tag := range tags {
			vwm, err := val.Properties(tag)

			if err != nil {
				e.client.logger.Debug(fmt.Sprintf("[ngorm] get properties by tag: %s, err: %v", tag, err))
				return nil, err
			}

			mt := make(map[string]any, len(vwm))

			for k := range vwm {
				data, err := e.handleValueWrapper(vwm[k])
				if err != nil {
					return nil, err
				}

				mt[k] = data
			}

			m[tag] = mt
		}

		return m, nil
	}

	return nil, ErrUnknownValueWrapper
}
