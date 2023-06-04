package ngorm

import (
	"errors"
)

var (
	ErrResultNil           = errors.New("result set is nil")
	ErrUnknownValueWrapper = errors.New("unknown nebula value_wrapper type")
	ErrNotPtr              = errors.New("scan object not ptr")
	ErrScanObject          = errors.New("scan object can't set")
	ErrScanObjectsLength   = errors.New("scan objects length error")
	ErrModelReflectInvalid = errors.New("model reflect invalid")
	ErrModelTypeUnsupport  = errors.New("model type unsupported")
	ErrSyntax              = errors.New("syntax error: %s")
)
