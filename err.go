package ngorm

import (
	"errors"
	"fmt"
)

var (
	ErrorResultNotFound   = errors.New("no row found in nebula")
	ErrorModelNotPtr      = errors.New("model type must be ptr")
	ErrorInvalidModel     = errors.New("model ptr invalid")
	ErrorInvalidModelType = errors.New("model type invalid")
	ErrorColumnNotFound   = errors.New("model column not found")
)

func ErrorColumnNotFoundGen(column string) error {
	return fmt.Errorf("%w: %s", ErrorColumnNotFound, column)
}

// ErrorUnknownNebulaValueType 无法处理的 nebula 数据类型
func ErrorUnknownNebulaValueType(t string) error {
	return fmt.Errorf("unknown nebula data type: %s", t)
}
