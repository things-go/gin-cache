package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/things-go/gin-cache/persist"
)

type RedisStore struct {
	Redisc *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client}
}

func (store *RedisStore) Set(key string, value interface{}, expire time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Redisc.Set(context.Background(), key, string(payload), expire).Err()
}

func (store *RedisStore) Get(key string, value interface{}) error {
	data, err := store.Redisc.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return persist.ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal(data, &value)
}

func (store *RedisStore) Delete(key string) error {
	return store.Redisc.Del(context.Background(), key).Err()
}
