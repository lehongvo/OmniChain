package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
)

// LocalCache implements in-memory cache using ristretto
type LocalCache struct {
	cache *ristretto.Cache
}

// NewLocalCache creates a new local cache instance
func NewLocalCache(maxCost int64, numCounters int64) (*LocalCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters, // Number of keys to track frequency
		MaxCost:     maxCost,     // Maximum cost of cache (in bytes)
		BufferItems: 64,          // Number of keys per Get buffer
	})
	if err != nil {
		return nil, err
	}

	return &LocalCache{cache: cache}, nil
}

// Get retrieves a value from cache
func (l *LocalCache) Get(ctx context.Context, key string) (string, error) {
	value, found := l.cache.Get(key)
	if !found {
		return "", ErrCacheMiss
	}

	str, ok := value.(string)
	if !ok {
		return "", ErrInvalidType
	}

	return str, nil
}

// Set sets a value in cache with TTL
func (l *LocalCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cost := int64(len(key) + estimateSize(value))
	success := l.cache.SetWithTTL(key, value, cost, ttl)
	if !success {
		return ErrCacheSetFailed
	}
	return nil
}

// Delete deletes a key from cache
func (l *LocalCache) Delete(ctx context.Context, key string) error {
	l.cache.Del(key)
	return nil
}

// Exists checks if a key exists
func (l *LocalCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := l.cache.Get(key)
	return found, nil
}

// Close closes the cache
func (l *LocalCache) Close() error {
	l.cache.Close()
	return nil
}

// estimateSize estimates the size of a value in bytes
func estimateSize(value interface{}) int {
	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	default:
		return 100 // Default estimate
	}
}

// Cache errors
var (
	ErrCacheMiss     = &CacheError{Message: "cache miss"}
	ErrInvalidType   = &CacheError{Message: "invalid cache value type"}
	ErrCacheSetFailed = &CacheError{Message: "failed to set cache value"}
)

// CacheError represents a cache error
type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

