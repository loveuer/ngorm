package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gitlab.umisen.com/tools/ngorm/v2"
	"os/signal"
	"syscall"
)

func main() {
	gctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	logrus.SetLevel(logrus.DebugLevel)

	client, err := ngorm.NewClient(gctx, &ngorm.Config{
		Endpoints:    []string{"10.220.10.48:9669"},
		Username:     "root",
		Password:     "123",
		DefaultSpace: "test_base",
		Logger:       nil,
	})

	if err != nil {
		logrus.Panic("init ngorm client err:", err)
	}

	app := gin.Default()
	app.GET("/uuid", func(c *gin.Context) {
		type Vertex struct {
			Id      string   `json:"id" nebula:"VertexID"`
			Names   []string `json:"names" nebula:"NAMES"`
			Address []string `json:"address" nebula:"ADDRESS"`
		}

		uuid := c.Query("uuid")
		if uuid == "" {
			c.JSON(400, "uuid is empty")
			return
		}

		result := new(Vertex)

		if err = client.Fetch(uuid).Tags("NAMES", "ADDRESS").Key("v").Scan(result); err != nil {
			c.JSON(500, err.Error())
			return
		}

		c.JSON(200, result)
	})

	logrus.Fatal(app.Run(":7777"))
}
