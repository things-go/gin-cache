# gin-cache gin's middleware

[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/things-go/gin-cache?tab=doc)
[![codecov](https://codecov.io/gh/things-go/gin-cache/branch/main/graph/badge.svg)](https://codecov.io/gh/things-go/gin-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/things-go/gin-cache)](https://goreportcard.com/report/github.com/things-go/gin-cache)
[![Licence](https://img.shields.io/github/license/things-go/gin-cache)](https://raw.githubusercontent.com/things-go/gin-cache/master/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/things-go/gin-cache)](https://github.com/things-go/gin-cache/tags)
[![Sourcegraph](https://sourcegraph.com/github.com/things-go/gin-cache/-/badge.svg)](https://sourcegraph.com/github.com/things-go/gin-cache?badge)

Gin middleware/handler to enable Cache.

## Usage

### Start using it

Download and install it:

```sh
go get github.com/things-go/gin-cache
```

Import it in your code:

```go
import cache "github.com/things-go/gin-cache"
```

### Example

#### 1. memory store

See the [memory store](_example/memory/memory.go)

[embedmd]:# (_example/memory/memory.go go)
```go
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

	app.GET("/hello",
		cache.CacheWithRequestURI(
			memory.NewStore(inmemory.New(time.Minute, time.Minute*10)),
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
```

#### 2. redis store

See the [redis store](_example/redis/redis.go)

[embedmd]:# (_example/redis/redis.go go)
```go
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

	store := redisStore.NewStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "localhost:6379",
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
```

#### 3. custom key 

See the [custom key](_example/custom/custom.go)

[embedmd]:# (_example/custom/custom.go go)
```go
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
```
