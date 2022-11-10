package main

import (
	"log"

	"github.com/loveuer/ngorm"
)

var (
	db  *ngorm.NGDB
	err error
)

func init() {
	db, err = ngorm.NewNGDB("test_base", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "10.220.10.19", Port: 9669}},
		Username: "root",
		Password: "xxx",
	})
	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}
}

func main() {
	type Vertex struct {
		ID      string   `nebula:"VertexID"` // Compatible vid
		Name    []string `nebula:"NAMES"`
		Address []string `nebula:"ADDRESS"`
		Company []string `nebula:"COMPANY"`
	}

	vertex := new(Vertex)
	vertexs := make([]Vertex, 0)
	vm := make(map[string]interface{})
	vms := make([]map[string]interface{}, 0)

	//f := db.Fetch([]string{"Bbp6S7"}...)
	f := db.Fetch([]string{"000164", "00031N6"}...).Key("v")
	f = f.Tags("*") // find all tags
	f = f.Tags()    // left tags with empty to find tags in model
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
		Name    string `nebula:"NAMES"`
		Company string `nebula:"COMPANY"`
	}

	vertexs := make([]Vertex, 0)
	if err = db.Fetch("vertex-1").Tags("NAME", "COMPANY").Key("name").Find(&vertexs); err != nil {
		log.Fatalln(err)
	}

	log.Println("vertexs:", vertexs)
}
