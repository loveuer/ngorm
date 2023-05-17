package main

import (
	"github.com/loveuer/ngorm"
	"github.com/sirupsen/logrus"
)

var (
	db *ngorm.NGDB
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)

	var (
		err   error
		space = "social_media_user3"
	)

	db, err = ngorm.NewNGDB(space, ngorm.Config{
		LogLevel: ngorm.DebugLevel,
		Servers: []ngorm.Service{
			{Addr: "10.230.200.201", Port: 9669},
		},
		Username: "root",
		Password: "xxx",
	})

	if err != nil {
		logrus.Fatalf("can't new ngdb with err: %v", err)
	}
}

func main() {
	type Result struct {
		CollectDate int         `nebula:"collect_date"`
		CreateDate  interface{} `nebula:"create_date"`
		Id          string      `nebula:"id"`
		Id2         string      `nebula:"id2"`
		InputDate   int         `nebula:"input_date"`
		Name        string      `nebula:"name"`
		Platform    string      `nebula:"platform"`
		Photo       string      `nebula:"photo"`
	}

	var (
		err  error
		list = make([]Result, 0)
	)

	if err = db.GOFrom("e0b5877c43f7c17e1d79fc1d4da9e44a").
		Over("like REVERSELY").
		Step(1).
		Yield("DISTINCT properties($$) as val").
		Finds(&list); err != nil {
		logrus.Fatalf("ngrom err: %v", err)
	}

	for i := range list {
		logrus.Printf("[idx: %d] result: %v", i, list[i])
	}
}
