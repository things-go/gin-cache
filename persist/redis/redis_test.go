package redis

import (
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/things-go/gin-cache/persist"
)

type cacheFactory func(*testing.T, time.Duration) persist.Store

// Test typical cache interactions
func typicalGetSet(t *testing.T, newCache cacheFactory) {
	var err error
	storeCache := newCache(t, time.Hour)

	value := "foo"
	err = storeCache.Set("value", value, time.Hour)
	require.NoError(t, err)

	value = ""
	err = storeCache.Get("value", &value)
	require.NoError(t, err)
	require.Equal(t, "foo", value)
}

func expiration(t *testing.T, newCache cacheFactory) {
	// memcached does not support expiration times less than 1 second.
	var err error
	storeCache := newCache(t, time.Second)

	value := 10

	// Test Set w/ short time
	err = storeCache.Set("int", value, time.Second)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
	err = storeCache.Get("int", &value)
	require.ErrorIs(t, err, persist.ErrCacheMiss)

	// Test Set w/ longer time.
	err = storeCache.Set("int", value, time.Hour)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
	err = storeCache.Get("int", &value)
	require.NoError(t, err)

	// Test Set w/ forever.
	err = storeCache.Set("int", value, -1)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)
	err = storeCache.Get("int", &value)
	require.NoError(t, err)
}

func emptyCache(t *testing.T, newCache cacheFactory) {
	var err error
	storeCache := newCache(t, time.Hour)

	err = storeCache.Get("notexist", time.Hour)
	require.Error(t, err)
	require.ErrorIs(t, err, persist.ErrCacheMiss)

	err = storeCache.Delete("notexist")
	require.NoError(t, err)
}

var newInRedisStore = func(_ *testing.T, defaultExpiration time.Duration) persist.Store {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return NewStore(redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + port,
	}))
}

// Test typical cache interactions
func Test_Memory_typicalGetSet(t *testing.T) {
	typicalGetSet(t, newInRedisStore)
}

func Test_Memory_Expiration(t *testing.T) {
	expiration(t, newInRedisStore)
}

func Test_Memory_Empty(t *testing.T) {
	emptyCache(t, newInRedisStore)
}
