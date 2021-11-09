package main

import (
	"gitee.com/loveuer/ngorm"
	"log"
)

func main() {
	db, err := ngorm.NewNGDB("Test", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "", Port: 8080}},
		Username: "admin",
		Password: "admin",
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

func example2() {
	db, err := ngorm.NewNGDB("Test", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "", Port: 8080}},
		Username: "admin",
		Password: "admin",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}

	vids := make([]string, 0)

	if err = db.Match("match (v1) -- (v2) where id(v1) == 'vertex-1' return id(v2)").Finds(&vids); err != nil {
		log.Fatalln(err)
	}

	log.Println("vids:", vids)
}
