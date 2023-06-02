package ngorm

import (
	"errors"
	"fmt"
)

var (
	ErrResultNil           = errors.New("result set is nil")
	ErrUnknownValueWrapper = errors.New("unknown nebula value_wrapper type")
	ErrNotPtr              = errors.New("object not ptr")
	ErrSyntax              = func(s string) error {
		return fmt.Errorf("syntax error: %s", s)
	}
)
