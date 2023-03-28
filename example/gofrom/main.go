package main

import (
	"fmt"
	"github.com/loveuer/ngorm"
	"log"
)

var (
	db *ngorm.NGDB
)

func init() {
	var (
		err   error
		space = ""
	)

	db, err = ngorm.NewNGDB(space, ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers: []ngorm.Service{
			{Addr: "", Port: 9669},
		},
		Username: "123",
		Password: "xxx",
	})

	if err != nil {
		log.Fatalf("can't new ngdb with err: %v", err)
	}
}

func main() {
	type Result struct {
		CollectDate int         `json:"collect_date" nebula:"collect_date"`
		CreateDate  int         `json:"create_date" nebula:"create_date"`
		Id          string      `json:"id" nebula:"id"`
		Id2         string      `json:"id2" nebula:"id2"`
		InputDate   int         `json:"input_date" nebula:"input_date"`
		Name        string      `json:"name" nebula:"name"`
		Photo       interface{} `json:"photo" nebula:"photo"`
		Platform    string      `json:"platform" nebula:"platform"`
	}

	var (
		err  error
		list = make([]Result, 0)
	)

	if err = db.GOFrom("twitter-whyyoutouzhele").
		Over("follow REVERSELY").
		Step(1).
		Yield(fmt.Sprintf("properties($$) as followers | limit %d, %d", 0, 10)).
		Finds(&list); err != nil {
		log.Fatalf("ngrom err: %v", err)
	}

	for i := range list {
		log.Printf("[idx: %d] result: %v", i, list[i])
	}
}
