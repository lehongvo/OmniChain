package metrics

import (
	"bytes"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// FiberMetricsHandler creates a Fiber handler for Prometheus metrics
func FiberMetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a buffer to capture metrics output
		var buf bytes.Buffer
		
		// Create a response writer that writes to buffer
		w := &bufferResponseWriter{
			buffer: &buf,
			header: make(http.Header),
		}
		
		// Create a simple HTTP request
		req, _ := http.NewRequest("GET", "/metrics", nil)
		
		// Serve metrics to buffer
		promhttp.Handler().ServeHTTP(w, req)
		
		// Set content type
		c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		
		// Write metrics to Fiber response
		return c.Send(buf.Bytes())
	}
}

type bufferResponseWriter struct {
	buffer *bytes.Buffer
	header http.Header
	status int
}

func (w *bufferResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferResponseWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *bufferResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

