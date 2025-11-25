package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/onichange/pos-system/pkg/metrics"
)

// PrometheusMetrics creates a middleware to record Prometheus metrics
func PrometheusMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Record metrics
		method := c.Method()
		endpoint := c.Route().Path
		if endpoint == "" {
			endpoint = c.Path()
		}
		statusCode := c.Response().StatusCode()

		metrics.RecordHTTPRequest(method, endpoint, statusCode, duration)

		return err
	}
}

