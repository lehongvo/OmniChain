package api

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

//go:embed openapi.yaml
var openAPIContent []byte

// SetupSwagger sets up Swagger UI for API documentation
func SetupSwagger(app *fiber.App) {
	// Use a dedicated group so we can relax security headers just for Swagger
	docs := app.Group("/api/docs")
	docs.Use(func(c *fiber.Ctx) error {
		// Allow inline scripts/styles and external fonts required by Swagger UI
		c.Set("Content-Security-Policy", ""+
			"default-src 'self' data: blob:; "+
			"img-src 'self' data:; "+
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
			"font-src 'self' data: https://fonts.gstatic.com; "+
			"connect-src 'self';")
		return c.Next()
	})

	// Serve OpenAPI spec file from embedded content
	docs.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		return c.Send(openAPIContent)
	})

	docs.Get("/doc.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		return c.Send(openAPIContent)
	})

	docs.Get("/doc.json", func(c *fiber.Ctx) error {
		// Swagger UI can read YAML, but if a client explicitly wants JSON we still return YAML
		c.Set("Content-Type", "application/x-yaml")
		return c.Send(openAPIContent)
	})

	// Swagger UI endpoint - MUST be last (wildcard route)
	// Access at: http://localhost:8080/api/docs/index.html
	docs.Get("/*", swagger.New(swagger.Config{
		URL:          "/api/docs/openapi.yaml",
		DeepLinking:  true,
		DocExpansion: "list",
		Title:        "OniChange POS System API",
	}))
}
