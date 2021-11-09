package main

import (
	"gitee.com/loveuer/ngorm"
	"log"
)

var (
	db *ngorm.NGDB
	err error
)

func init() {
	db, err = ngorm.NewNGDB("Test", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "", Port: 8080}},
		Username: "admin",
		Password: "admin",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}
}

func main() {
	type Vertex struct {
		ID      string `nebula:"VertexID"`
		Name    string `nebula:"NAME"`
		Company string `nebula:"COMPANY"`
	}

	vertex := new(Vertex)

	if err = db.Fetch("vertex-1").Tags("NAME", "COMPANY").Key("name").Find(vertex); err != nil {
		log.Fatalln(err)
	}

	log.Println("vertex:", vertex)
}

func example2() {
	type Vertex struct {
		ID      string `nebula:"VertexID"`
		Name    string `nebula:"NAME"`
		Company string `nebula:"COMPANY"`
	}

	vertexs := make([]Vertex, 0)
	if err = db.Fetch("vertex-1").Tags("NAME", "COMPANY").Key("name").Find(&vertexs); err != nil {
		log.Fatalln(err)
	}

	log.Println("vertexs:", vertexs)
}
