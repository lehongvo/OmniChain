package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/singleflight"
)

// RequestCoalescingMiddleware coalesces duplicate requests
type RequestCoalescingMiddleware struct {
	group singleflight.Group
	mu    sync.RWMutex
	cache map[string]*cachedResponse
}

type cachedResponse struct {
	data      []byte
	timestamp time.Time
	ttl       time.Duration
}

// NewRequestCoalescingMiddleware creates a new request coalescing middleware
func NewRequestCoalescingMiddleware() *RequestCoalescingMiddleware {
	return &RequestCoalescingMiddleware{
		cache: make(map[string]*cachedResponse),
	}
}

// Middleware returns the middleware handler
func (rcm *RequestCoalescingMiddleware) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only coalesce GET requests
		if c.Method() != "GET" {
			return c.Next()
		}

		// Create cache key from request
		queries := c.Queries()
		queryStr := ""
		if len(queries) > 0 {
			queryStr = "?"
			first := true
			for k, v := range queries {
				if !first {
					queryStr += "&"
				}
				queryStr += k + "=" + v
				first = false
			}
		}
		cacheKey := c.Path() + queryStr

		// Check cache first
		rcm.mu.RLock()
		if cached, ok := rcm.cache[cacheKey]; ok {
			if time.Since(cached.timestamp) < cached.ttl {
				rcm.mu.RUnlock()
				c.Set("X-Cache", "HIT")
				return c.Send(cached.data)
			}
		}
		rcm.mu.RUnlock()

		// Use singleflight to coalesce duplicate requests
		result, err, shared := rcm.group.Do(cacheKey, func() (interface{}, error) {
			// Process request
			if err := c.Next(); err != nil {
				return nil, err
			}

			// Get response
			response := c.Response().Body()
			return response, nil
		})

		if err != nil {
			return err
		}

		// If request was shared (coalesced), set header
		if shared {
			c.Set("X-Request-Coalesced", "true")
		}

		// Cache response (with short TTL for GET requests)
		rcm.mu.Lock()
		rcm.cache[cacheKey] = &cachedResponse{
			data:      result.([]byte),
			timestamp: time.Now(),
			ttl:       5 * time.Second, // Short TTL for request coalescing
		}
		rcm.mu.Unlock()

		return c.Send(result.([]byte))
	}
}
