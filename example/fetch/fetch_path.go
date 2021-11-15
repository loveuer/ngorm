package main

import (
	"gitee.com/loveuer/ngorm"
	"log"
)

func fetchPath() {
	fdb, err := ngorm.NewNGDB("space", ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers:  []ngorm.Service{{Addr: "192.168.1.1", Port: 80}},
		Username: "user",
		Password: "nebula",
	})
	if err != nil {
		log.Panicln(err)
	}

	resp := make([]map[string]interface{}, 0)

	type CName struct {
		Name []string `nebula:"name"`
	}

	ns := make([]CName, 0)

	fdb.FetchPath("edge_type").Path(`"vertex-1" -> "vertex-2"`, `"vertex-9" -> "vertex-9119"`).Props("name").Find(&ns)

	log.Println(resp)
	log.Println("ns:", ns)
}
