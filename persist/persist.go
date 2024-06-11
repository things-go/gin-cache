package persist

import (
	"errors"
	"time"
)

// ErrCacheMiss cache miss error
var ErrCacheMiss = errors.New("persist: cache miss")

// Store is the interface of a Cache backend
type Store interface {
	// Get retrieves an item from the Cache. Returns the item or nil, and a bool indicating
	// whether the key was found.
	Get(key string, value any) error

	// Set sets an item to the Cache, replacing any existing item.
	Set(key string, value any, expire time.Duration) error

	// Delete removes an item from the Cache. Does nothing if the key is not in the Cache.
	Delete(key string) error
}
