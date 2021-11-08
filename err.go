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
)

func ErrorColumnNotFound(column string) error {
	return fmt.Errorf("model column: %s not found", column)
}

// ErrorUnknownNebulaValueType 无法处理的 nebula 数据类型
func ErrorUnknownNebulaValueType(t string) error {
	return fmt.Errorf("unknown nebula data type: %s", t)
}
