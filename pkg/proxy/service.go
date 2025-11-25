package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ServiceProxy handles proxying requests to microservices
type ServiceProxy struct {
	client  *http.Client
	baseURL string
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(baseURL string) *ServiceProxy {
	return &ServiceProxy{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
	}
}

// Proxy proxies the request to the target service
func (p *ServiceProxy) Proxy(c *fiber.Ctx) error {
	// Build target URL
	targetURL := p.baseURL + c.Path()

	// Add query parameters
	queries := c.Queries()
	if len(queries) > 0 {
		values := url.Values{}
		for key, value := range queries {
			values.Set(key, value)
		}
		if queryString := values.Encode(); queryString != "" {
			targetURL += "?" + queryString
		}
	}

	// Create request
	req, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "Failed to create request")
	}

	// Copy headers (except Host and some Fiber-specific headers)
	skipHeaders := map[string]bool{
		"Host":           true,
		"Connection":     true,
		"Content-Length": true,
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := strings.ToLower(string(key))
		if !skipHeaders[keyStr] {
			req.Header.Add(string(key), string(value))
		}
	})

	// Forward user information
	if userID := c.Locals("user_id"); userID != nil {
		req.Header.Set("X-User-ID", userID.(string))
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "Failed to connect to service")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "Failed to read response")
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set status and body
	c.Status(resp.StatusCode)
	return c.Send(body)
}
