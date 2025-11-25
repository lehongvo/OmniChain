package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/onichange/pos-system/pkg/config"
	"github.com/onichange/pos-system/pkg/logger"
	"github.com/onichange/pos-system/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Server.Environment)
	log.Info("Starting Payment Service...")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorHandler: errorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.RequestSizeLimit(cfg.Security.MaxRequestSize))
	app.Use(middleware.PrometheusMetrics())

	if cfg.Security.EnableCORS {
		app.Use(middleware.CORSMiddleware(cfg.Security.CORSOrigins))
	}

	// Health check endpoints
	app.Get("/health", healthCheck)
	app.Get("/ready", readinessCheck)

	// API routes
	api := app.Group("/api/v1")

	// Payment routes (to be implemented - PCI-DSS compliant)
	api.Post("/payments", processPayment)
	api.Get("/payments/:id", getPaymentByID)
	api.Post("/payments/:id/refund", refundPayment)
	api.Get("/payments/:id/status", getPaymentStatus)
	api.Post("/payments/verify", verifyPayment)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, "8084") // Payment service port

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Infof("Payment Service listening on %s", addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Payment Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	log.Info("Payment Service stopped")
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"service":   "payment-service",
		"timestamp": time.Now().Unix(),
	})
}

func readinessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ready",
		"service":   "payment-service",
		"timestamp": time.Now().Unix(),
	})
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

// Placeholder handlers - to be implemented with PCI-DSS compliance
func processPayment(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Process payment - to be implemented (PCI-DSS compliant)"})
}

func getPaymentByID(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get payment by ID - to be implemented"})
}

func refundPayment(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Refund payment - to be implemented"})
}

func getPaymentStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get payment status - to be implemented"})
}

func verifyPayment(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Verify payment - to be implemented"})
}
