package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ETagMiddleware adds ETag support for HTTP caching
func ETagMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only cache GET requests
		if c.Method() != "GET" {
			return c.Next()
		}

		// Generate ETag from response
		originalBody := c.Response().Body()

		// Process request first
		if err := c.Next(); err != nil {
			return err
		}

		// Generate ETag from response body
		body := c.Response().Body()
		hash := sha256.Sum256(body)
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		// Set ETag header
		c.Set("ETag", etag)

		// Check If-None-Match header
		if match := c.Get("If-None-Match"); match == etag {
			c.Status(fiber.StatusNotModified)
			return c.Send(nil)
		}

		// Restore original body if needed
		if len(originalBody) > 0 {
			c.Response().SetBody(originalBody)
		}

		return nil
	}
}

// CacheControlMiddleware adds Cache-Control headers
func CacheControlMiddleware(maxAge time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
		return c.Next()
	}
}
