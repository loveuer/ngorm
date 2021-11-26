package main

import (
	"gitee.com/loveuer/ngorm"
	"log"
)

var (
	db  *ngorm.NGDB
	err error
)

func init() {
	db, err = ngorm.NewNGDB("space_name", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "127.0.0.1", Port: 9669}},
		Username: "xxx",
		Password: "xxx",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}
}

func main() {
	type Vertex struct {
		ID      string   `nebula:"VertexID"`
		Name    []string `nebula:"Name"`
		Address []string `nebula:"Address"`
		Company []string `nebula:"Company"`
	}

	vertex := new(Vertex)
	vertexs := make([]Vertex, 0)
	vm := make(map[string]interface{})
	vms := make([]map[string]interface{}, 0)

	//f := db.Fetch([]string{"Bbp6S7"}...)
	f := db.Fetch([]string{"vertex-1", "vertex-2"}...)
	//err = f.Find(&vm)
	//err = f.Find(vertex)
	err = f.Find(&vertexs)
	//err = f.Find(&vms)
	if err != nil {
		log.Printf("err: %v", err)
	}

	log.Printf("vertex:  %v", vertex)
	log.Printf("v map:   %v", vm)
	log.Printf("vertexs: %v", vertexs)
	log.Printf("v maps : %v", vms)
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
