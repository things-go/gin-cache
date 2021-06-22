package main

import (
	"time"

	"github.com/gin-gonic/gin"
	mcache "github.com/patrickmn/go-cache"

	cache "github.com/things-go/gin-cache"
	"github.com/things-go/gin-cache/persist/memory"
)

func main() {
	app := gin.New()

	app.GET("/hello",
		cache.CacheWithRequestURI(
			memory.NewStore(mcache.New(time.Minute, time.Minute*10)),
			5*time.Second,
			func(c *gin.Context) {
				c.String(200, "hello world")
			},
		),
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
