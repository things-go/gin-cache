# gin-cache gin's middleware

[![GoDoc](https://godoc.org/github.com/things-go/gin-cache?status.svg)](https://godoc.org/github.com/things-go/gin-cache)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/things-go/gin-cache?tab=doc)
[![Build Status](https://www.travis-ci.com/things-go/gin-cache.svg?branch=master)](https://www.travis-ci.com/things-go/gin-cache)
[![codecov](https://codecov.io/gh/things-go/gin-cache/branch/master/graph/badge.svg)](https://codecov.io/gh/things-go/gin-cache)
![Action Status](https://github.com/things-go/gin-cache/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/things-go/gin-cache)](https://goreportcard.com/report/github.com/things-go/gin-cache)
[![Licence](https://img.shields.io/github/license/things-go/gin-cache)](https://raw.githubusercontent.com/things-go/gin-cache/master/LICENSE)
[![Tag](https://img.shields.io/github/v/tag/things-go/gin-cache)](https://github.com/things-go/gin-cache/tags)
[![Sourcegraph](https://sourcegraph.com/github.com/things-go/gin-cache/-/badge.svg)](https://sourcegraph.com/github.com/things-go/gin-cache?badge)

Gin middleware/handler to enable Cache.

## Usage

### Start using it

Download and install it:

```sh
$ go get github.com/things-go/gin-cache
```

Import it in your code:

```go
import cache "github.com/things-go/gin-cache"
```

### Example:

#### 1. memory store

See the [memory store](_example/memory/memory.go)

[embedmd]:# (_example/memory/memory.go go)
```go
package main

import (
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
```
### Donation

if package help you a lot,you can support us by:

**Alipay**

![alipay](https://github.com/thinkgos/thinkgos/blob/master/asserts/alipay.jpg)

**WeChat Pay**

![wxpay](https://github.com/thinkgos/thinkgos/blob/master/asserts/wxpay.jpg)
