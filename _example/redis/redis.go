package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	cache "github.com/things-go/gin-cache"
	redisStore "github.com/things-go/gin-cache/persist/redis"
)

func main() {
	app := gin.New()

	store := redisStore.NewRedisStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	}))

	app.GET("/hello",
		cache.CacheWithRequestPath(
			store,
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
