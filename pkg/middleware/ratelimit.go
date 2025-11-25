package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiter implements token bucket algorithm with Redis
type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// RateLimitMiddleware returns a Fiber middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP address or user ID)
		identifier := c.IP()
		if userID := c.Get("X-User-ID"); userID != "" {
			identifier = userID
		}

		key := fmt.Sprintf("ratelimit:%s", identifier)
		ctx := context.Background()

		// Use Redis sliding window log algorithm
		now := time.Now().Unix()
		windowStart := now - int64(rl.window.Seconds())

		// Remove old entries
		rl.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// Count current requests
		count, err := rl.client.ZCard(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow request (fail open)
			return c.Next()
		}

		if count >= int64(rl.limit) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
		}

		// Add current request
		rl.client.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d", now),
		})
		rl.client.Expire(ctx, key, rl.window)

		return c.Next()
	}
}

