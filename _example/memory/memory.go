package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	cache "github.com/things-go/gin-cache"
	"github.com/things-go/gin-cache/persist/memory"
)

func main() {
	app := gin.New()

	app.GET("/hello",
		cache.CacheWithRequestURI(
			memory.NewMemoryStore(1*time.Minute),
			5*time.Second,
			func(c *gin.Context) {
				log.Println(c.Request.URL.RequestURI())
				log.Println(c.Request.URL.Path)
				c.String(200, "hello world")
			},
		),
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
