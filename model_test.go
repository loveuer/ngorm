package ngorm

import (
	"testing"
)

func TestParse(t *testing.T) {
	type User struct {
		Id       uint64   `nebula:"id.v"`
		Username string   `nebula:"username.v"`
		Likes    []string `nebula:"likes.v"`
	}

	var (
		a any
	)

	// test struct
	if m, err := Parse(&User{}); err != nil {
		t.Error("test struct err:", err)
		_ = m
	}

	// test []struct
	us := make([]*User, 0)
	if m, err := Parse(&us); err != nil {
		t.Error("test []struct err:", err)
		_ = m
	}

	// test any
	if m, err := Parse(&a); err != nil {
		t.Error("test any err:", err)
		_ = m
	}

	// test map[string]any
	um := make(map[string]any)
	if m, err := Parse(&um); err != nil {
		t.Error("test map[string]any err:", err)
		_ = m
	}

	// test []map[string]any
	ums := make([]map[string]any, 0)
	if m, err := Parse(&ums); err != nil {
		t.Error("test []map[string]any err:", err)
		_ = m
	}
}
