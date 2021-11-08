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
	_ = db
}
