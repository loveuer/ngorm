package ngorm

import "errors"

var (
	ErrResultNil           = errors.New("result set is nil")
	ErrUnknownValueWrapper = errors.New("unknown nebula value_wrapper type")
)
