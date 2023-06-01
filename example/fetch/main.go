package main

import (
	"github.com/sirupsen/logrus"
	"gitlab.umisen.com/tools/ngorm/v2"
)

func main() {
	client, err := ngorm.NewClient(&ngorm.Config{
		Endpoints:    []string{"10.220.10.19:9669"},
		Username:     "root",
		Password:     "123",
		DefaultSpace: "test_base",
		Logger:       nil,
	})

	if err != nil {
		logrus.Panic(err)
	}

	var v any

	if err = client.Raw("match (v1) return v1 limit 5").Scan(&v); err != nil {
		logrus.Panic("scan err:", err)
	}
}
