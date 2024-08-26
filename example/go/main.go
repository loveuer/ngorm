package main

import (
	"fmt"
	"github.com/loveuer/ngorm/v2"
)

func main() {}

var (
	client *ngorm.Client
)

func GetTargetVertexID() {
	// todo: init client first

	type Res struct {
		Target string `nebula:"VertexID"`
	}

	result := make([]*Res, 0)

	if err := client.Session().GoFrom("source_id").Steps(1).Over("edge_name").Scan(&result); err != nil {
		panic(err)
	}

	for idx := range result {
		fmt.Printf("target => %s\n", result[idx].Target)
	}
}
