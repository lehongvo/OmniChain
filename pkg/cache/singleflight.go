package cache

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// SingleFlightCache wraps cache with singleflight to prevent cache stampede
type SingleFlightCache struct {
	cache      Cache
	group      singleflight.Group
	mu         sync.RWMutex
	expiration map[string]time.Time
}

// Cache interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewSingleFlightCache creates a new cache with singleflight protection
func NewSingleFlightCache(cache Cache) *SingleFlightCache {
	return &SingleFlightCache{
		cache:      cache,
		expiration: make(map[string]time.Time),
	}
}

// GetWithFallback gets a value from cache, or calls fallback if not found
// Prevents cache stampede by using singleflight
func (sfc *SingleFlightCache) GetWithFallback(
	ctx context.Context,
	key string,
	ttl time.Duration,
	fallback func() (string, error),
) (string, error) {
	// Check cache first
	value, err := sfc.cache.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// Use singleflight to prevent multiple concurrent calls
	result, err, _ := sfc.group.Do(key, func() (interface{}, error) {
		// Double-check cache (another goroutine might have set it)
		value, err := sfc.cache.Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// Call fallback function
		value, err = fallback()
		if err != nil {
			return nil, err
		}

		// Set in cache
		if setErr := sfc.cache.Set(ctx, key, value, ttl); setErr != nil {
			// Log error but don't fail
		}

		return value, nil
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// Get retrieves a value from cache
func (sfc *SingleFlightCache) Get(ctx context.Context, key string) (string, error) {
	return sfc.cache.Get(ctx, key)
}

// Set sets a value in cache
func (sfc *SingleFlightCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	sfc.mu.Lock()
	sfc.expiration[key] = time.Now().Add(ttl)
	sfc.mu.Unlock()
	return sfc.cache.Set(ctx, key, value, ttl)
}

// Delete deletes a key from cache
func (sfc *SingleFlightCache) Delete(ctx context.Context, key string) error {
	sfc.mu.Lock()
	delete(sfc.expiration, key)
	sfc.mu.Unlock()
	return sfc.cache.Delete(ctx, key)
}

// Exists checks if a key exists
func (sfc *SingleFlightCache) Exists(ctx context.Context, key string) (bool, error) {
	return sfc.cache.Exists(ctx, key)
}
