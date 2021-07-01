package main

import (
	"time"

	"github.com/gin-gonic/gin"
	inmemory "github.com/patrickmn/go-cache"

	cache "github.com/things-go/gin-cache"
	"github.com/things-go/gin-cache/persist/memory"
)

func main() {
	app := gin.New()

	app.GET("/hello/:a/:b", custom())
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}

func custom() gin.HandlerFunc {
	f := cache.CacheWithRequestURI(
		memory.NewStore(inmemory.New(time.Minute, time.Minute*10)),
		5*time.Second,
		func(c *gin.Context) {
			c.String(200, "hello world")
		},
		cache.WithGenerateKey(func(c *gin.Context) (string, bool) {
			return c.GetString("custom_key"), true
		}),
	)
	return func(c *gin.Context) {
		a := c.Param("a")
		b := c.Param("b")
		c.Set("custom_key", cache.GenerateKeyWithPrefix(cache.PageCachePrefix, a+":"+b))
		f(c)
	}
}
