package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/things-go/gin-cache/persist"
)

// Store redis store
type Store struct {
	Redisc *redis.Client
}

// NewRedisStore new redis store
func NewRedisStore(client *redis.Client) *Store {
	return &Store{client}
}

// Set implement persist.Store interface
func (store *Store) Set(key string, value interface{}, expire time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Redisc.Set(context.Background(), key, string(payload), expire).Err()
}

// Get implement persist.Store interface
func (store *Store) Get(key string, value interface{}) error {
	data, err := store.Redisc.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return persist.ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal(data, &value)
}

// Delete implement persist.Store interface
func (store *Store) Delete(key string) error {
	return store.Redisc.Del(context.Background(), key).Err()
}
