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
		DefaultSpace: "test_base",
		Logger:       nil,
	})
	var (
		v1 = make([]string, 0)
		v2 = make([]string, 0)
	)
	var count int64
	v1 = append(v1, "4m6ziH3")
	if err := client.Match(&v1, "head").With(&v2, "v2", ForwardDirection).Where("id(head)", v1).Select("head", "headv2", "v2").Count(&count); err != nil {
		fmt.Printf("err:%v\n", err)
	}
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
	type FocusListRes struct {
		Uuid             string   `nebula:"VertexID" json:"uuid"`
		Email            []string `nebula:"EMAIL" json:"email"`
		RelationCountDST []string `nebula:"RELATION_COUNT_DST" json:"relation_count_dst"`
		Region           []string `nebula:"ADDRESS" json:"region"`
		Photo            []string `nebula:"PHOTO" json:"photo"`
		Phone            []string `nebula:"PHONE" json:"phone"`
		Names            []string `nebula:"NAMES" json:"names"`
	}
	type contact struct {
		Edge  string   `nebula:"edge"`
		Src   string   `nebula:"src"`
		Dst   string   `nebula:"dst"`
		Rank  int64    `nebula:"rank"`
		Names []string `nebula:"names"`
	}
	var (
		v1 = make([]string, 0)
		v2 = make([]string, 0)
	)
	users := make([]contact, 0)
	users2 := make([]FocusListRes, 0)
	v1 = append(v1, "4m6ziH3")
	if err := client.Match(&v1, "v1").With(&v2, "v2", ForwardDirection).Key("v").Where("id(v1)", v1).Select("v1v2", "v2").Limit(10).Finds(&users2, &users); err != nil {
		fmt.Printf("err:%v\n", err)
	}
	for i := range users {
		fmt.Printf("user-------%v:%+v\n", i, users[i])
	}
	for i := range users2 {
		fmt.Printf("users2-------%v:%+v\n", i, users2[i])
	}
}
