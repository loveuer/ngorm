package main

import (
	"log"

	"gitee.com/loveuer/ngorm"
	"github.com/ysmood/got/lib/gop"
)

func example1() {
	db, err := ngorm.NewNGDB("test_base", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "10.220.10.19", Port: 9669}},
		Username: "root",
		Password: "xxx",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}

	type Vertex struct {
		ID      string   `nebula:"VertexID"`
		Name    []string `nebula:"NAME"`
		Company []string `nebula:"COMPANY"`
	}

	src := new(Vertex)
	dsts := make([]Vertex, 0)

	err = db.Match("match (v1) - [e:edge_type] - (v2) where id(v1) == 'vertex-1' return v1, v2").Finds(src, &dsts)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("src:", src, "dsts:", dsts)
}

func main() {
	db, err := ngorm.NewNGDB("test_base_organization", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "10.220.10.19", Port: 9669}},
		Username: "root",
		Password: "xxx",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}

	type User struct {
		ID    string   `nebula:"VertexID"`
		Names []string `nebula:"names"`
	}

	type Org struct {
		ID    string   `nebula:"VertexID"`
		Names []string `nebula:"names"`
	}

	u1 := make([]User, 0)
	u2 := make([]User, 0)
	o3 := make([]Org, 0)

	if err = db.Match(
		"match (v:USER_INFO)--(v1:USER_INFO)--(v3:ORG_INFO) where id(v) == '4U2izs' return v,v1, v3 limit 3",
	).
		Finds(&u1, &u2, &o3); err != nil {
		log.Fatalln(err)
	}

	gop.P("user1:\n", u1)
	gop.P("user2:\n", u2)
	gop.P("org3:\n", o3)
}
