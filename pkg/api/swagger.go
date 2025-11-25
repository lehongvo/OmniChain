package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupSwagger sets up Swagger UI for API documentation
func SetupSwagger(app *fiber.App) {
	// Swagger UI endpoint
	app.Get("/api/docs/*", swagger.HandlerDefault)
}

