package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupSwagger sets up Swagger UI for API documentation
func SetupSwagger(app *fiber.App) {
	// Swagger UI endpoint
	// Access at: http://localhost:8080/api/docs/index.html
	// Swagger will automatically serve the OpenAPI spec from docs/api/openapi.yaml
	app.Get("/api/docs/*", swagger.New(swagger.Config{
		URL:          "/api/docs/doc.json",
		DeepLinking:  true,
		DocExpansion: "list",
		Title:        "OniChange POS System API",
	}))
}
