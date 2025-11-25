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

	"github.com/onichange/pos-system/internal/infrastructure/repository"
	"github.com/onichange/pos-system/internal/interfaces/http/store"
	"github.com/onichange/pos-system/pkg/config"
	"github.com/onichange/pos-system/pkg/database"
	"github.com/onichange/pos-system/pkg/logger"
	"github.com/onichange/pos-system/pkg/metrics"
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
	log.Info("Starting Store Service...")

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database, log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	storeRepo := repository.NewStoreRepository(db.Pool)

	// Initialize handlers
	storeHandler := store.NewHandler(storeRepo)

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

	// Prometheus metrics endpoint
	app.Get("/metrics", metrics.FiberMetricsHandler())

	// API routes
	api := app.Group("/api/v1")

	// Store routes
	api.Get("/stores", storeHandler.GetStores)
	api.Get("/stores/search", storeHandler.SearchStores)
	api.Get("/stores/:id", storeHandler.GetStoreByID)
	api.Post("/stores", storeHandler.CreateStore)
	api.Put("/stores/:id", storeHandler.UpdateStore)
	api.Delete("/stores/:id", storeHandler.DeleteStore)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, "8083") // Store service port

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Infof("Store Service listening on %s", addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Store Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	log.Info("Store Service stopped")
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"service":   "store-service",
		"timestamp": time.Now().Unix(),
	})
}

func readinessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ready",
		"service":   "store-service",
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
