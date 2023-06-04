package main

import (
	"github.com/gin-gonic/gin"
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

	// fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address

	app := gin.Default()

	app.GET("/data", func(c *gin.Context) {
		result, err := client.Raw("GO 1 steps FROM '4m6ziH3' OVER contact YIELD contact._dst AS destination").Result()
		if err != nil {
			c.JSON(500, err.Error())
			c.Abort()
			return
		}

		c.JSON(200, result)
	})

	app.Run(":7667")
}
