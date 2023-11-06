package ngorm

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestReflect(t *testing.T) {
	type UserInfo struct {
		Limit  *int
		Offset int
	}
	sc := make([]string, 0)
	typeof := reflect.TypeOf(&sc)

	switch typeof.Kind() {
	case reflect.Struct:
		fmt.Println("sc is struct")

	default:
		fmt.Println("sc not is struct")
	}
	var (
		tableName string
	)
	tableName = typeof.Name()[0:1]
	for _, v := range typeof.Name()[1:] {
		if v >= 65 && v <= 90 {
			tableName += "_" + string(v)
		} else if v >= 97 && v <= 122 {
			tableName += string(v - 32)
		} else {
			tableName += string(v)
		}
	}
	fmt.Println(tableName)
}

func TestCount(t *testing.T) {
	client, _ := NewClient(context.TODO(), &Config{
		Endpoints:    []string{"10.220.10.19:9669"},
		Username:     "root",
		Password:     "123",
		DefaultSpace: "test_base_organization",
		Logger:       nil,
	})
	var (
		v1 = make([]string, 0)
		v2 = make([]string, 0)
	)
	var count int64
	v1 = append(v1, "4m6ziH3")
	client.MatchHead(&v1).With(TwoDirection, &v2).CountPath(&count)
	fmt.Println("count:", count)
}

func TestFind(t *testing.T) {
	client, _ := NewClient(context.TODO(), &Config{
		Endpoints:    []string{"10.220.10.19:9669"},
		Username:     "root",
		Password:     "123",
		DefaultSpace: "test_base",
		Logger:       nil,
	})
	type edge struct {
		Edge  string   `nebula:"edge"`
		Names []string `nebula:"names"`
	}

	var (
		v1 = make([]string, 0)
		v2 = make([]string, 0)
	)
	users := make([]edge, 0)
	v1 = append(v1, "4m6ziH3")
	client.MatchHead(&v1).With(ForwardDirection, &v2).Limit(10).FindPath(&users)
	fmt.Println("users:", users)
}
