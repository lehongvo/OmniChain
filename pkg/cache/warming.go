package cache

import (
	"context"
	"time"

	"github.com/onichange/pos-system/pkg/logger"
)

// CacheWarmer handles cache warming on startup
type CacheWarmer struct {
	cache  Cache
	logger *logger.Logger
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(cache Cache, log *logger.Logger) *CacheWarmer {
	return &CacheWarmer{
		cache:  cache,
		logger: log,
	}
}

// WarmupItem represents an item to warm up
type WarmupItem struct {
	Key      string
	GetValue func() (interface{}, error)
	TTL      time.Duration
}

// Warmup warms up the cache with provided items
func (cw *CacheWarmer) Warmup(ctx context.Context, items []WarmupItem) error {
	cw.logger.Info("Starting cache warmup...")

	for _, item := range items {
		// Check if already exists
		exists, err := cw.cache.Exists(ctx, item.Key)
		if err == nil && exists {
			continue
		}

		// Get value
		value, err := item.GetValue()
		if err != nil {
			cw.logger.Warnf("Failed to warmup cache key %s: %v", item.Key, err)
			continue
		}

		// Set in cache
		if err := cw.cache.Set(ctx, item.Key, value, item.TTL); err != nil {
			cw.logger.Warnf("Failed to set cache key %s: %v", item.Key, err)
			continue
		}

		cw.logger.Debugf("Warmed up cache key: %s", item.Key)
	}

	cw.logger.Info("Cache warmup completed")
	return nil
}
