package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/loveuer/esgo2dump/log"
	"github.com/loveuer/ngorm/v2"
	"net/http"
	"os/signal"
	"syscall"
)

type Vertex struct {
	Id      string   `json:"id" nebula:"VertexID"`
	Names   []string `json:"names" nebula:"NAMES"`
	Address []string `json:"address" nebula:"ADDRESS"`
}

func main() {
	gctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	log.SetLogLevel(log.LogLevelDebug)

	client, err := ngorm.NewClient(gctx, &ngorm.Config{
		Endpoints:    []string{"localhost.nebula:9669"},
		Username:     "root",
		Password:     "password",
		DefaultSpace: "test_base",
		Logger:       nil,
	})

	if err != nil {
		log.Panic("init ngorm client err: %v", err)
	}

	app := gin.Default()
	app.GET("/fetch", func(c *gin.Context) {
		uuid := c.Query("uuid")
		if uuid == "" {
			c.JSON(400, "uuid is empty")
			return
		}

		result := new(Vertex)

		if err = client.Session().Fetch(uuid).Tags("NAMES", "ADDRESS").Key("v").Scan(result); err != nil {
			c.JSON(500, err.Error())
			return
		}

		c.JSON(200, result)
	})

	app.GET("/go", func(c *gin.Context) {
		uuid := c.Query("uuid")

		if uuid == "" {
			c.JSON(400, "uuid is empty")
			return
		}

		results := make([]*Vertex, 0)

		if err = client.GoFrom(uuid).
			Model(&Vertex{}).
			Over("relation", ngorm.EdgeTypeBoth).
			Tags("NAMES", "ADDRESS").
			Key("v").
			Scan(&results); err != nil {
			c.JSON(500, err.Error())
			return
		}

		c.JSON(200, results)
	})

	srv := http.Server{Handler: app, Addr: "0.0.0.0:7788"}

	go func() {
		log.Fatal(srv.ListenAndServe().Error())
	}()

	go func(ctx context.Context) {
		<-ctx.Done()
		_ = srv.Shutdown(context.TODO())
	}(gctx)

	<-gctx.Done()
}
