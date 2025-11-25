package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

// EnablePprof enables pprof endpoints for profiling
func EnablePprof(app *fiber.App) {
	// Only enable in development or with proper authentication
	app.Use("/debug/pprof", pprof.New())
}

