package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/onichange/pos-system/pkg/api"
	"github.com/onichange/pos-system/pkg/auth"
	"github.com/onichange/pos-system/pkg/cache"
	"github.com/onichange/pos-system/pkg/config"
	"github.com/onichange/pos-system/pkg/logger"
	"github.com/onichange/pos-system/pkg/middleware"
	"github.com/onichange/pos-system/pkg/proxy"
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
	log.Info("Starting API Gateway...")

	// Initialize Redis for rate limiting
	redisCache, err := cache.NewRedisCache(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Redis.MinIdleConns,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Get Redis client for rate limiter
	redisClient := redisCache.GetClient()

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		redisClient,
		cfg.Security.RateLimitRequestsPerMinute,
		time.Minute,
	)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)

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
	app.Use(middleware.PrometheusMetrics()) // Prometheus metrics

	if cfg.Security.EnableCORS {
		app.Use(middleware.CORSMiddleware(cfg.Security.CORSOrigins))
	}

	// Rate limiting middleware
	app.Use(rateLimiter.RateLimitMiddleware())

	// Start Prometheus metrics server on separate port
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{
			Addr:    ":9090",
			Handler: metricsMux,
		}
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Metrics server error: %v", err)
		}
	}()
	log.Info("Prometheus metrics server started on :9090/metrics")

	// pprof endpoints (only in development)
	if cfg.Server.Environment == "development" {
		app.Use("/debug/pprof", pprof.New())
	}

	// Swagger/OpenAPI documentation
	// Serve OpenAPI spec file
	app.Static("/api/docs", "./docs/api")
	// Setup Swagger UI
	api.SetupSwagger(app)

	// Health check endpoint
	app.Get("/health", healthCheck)
	app.Get("/ready", readinessCheck)

	// API routes
	api := app.Group("/api/v1")

	// Authentication routes
	authGroup := api.Group("/auth")
	authGroup.Post("/login", handleLogin(jwtManager))
	authGroup.Post("/refresh", handleRefresh(jwtManager))
	authGroup.Post("/logout", handleLogout)

	// Protected routes with JWT authentication
	protected := api.Group("/", middleware.JWTAuth(jwtManager))

	// Order service routes
	orderProxy := proxy.NewServiceProxy(cfg.Services.OrderServiceURL)
	protected.Get("/orders", orderProxy.Proxy)
	protected.Post("/orders", orderProxy.Proxy)
	protected.Get("/orders/:id", orderProxy.Proxy)
	protected.Put("/orders/:id", orderProxy.Proxy)
	protected.Delete("/orders/:id", orderProxy.Proxy)

	// User service routes
	userProxy := proxy.NewServiceProxy(cfg.Services.UserServiceURL)
	protected.Get("/users/me", userProxy.Proxy)
	protected.Put("/users/me", userProxy.Proxy)

	// Store service routes
	storeProxy := proxy.NewServiceProxy(cfg.Services.StoreServiceURL)
	protected.Get("/stores", storeProxy.Proxy)
	protected.Get("/stores/:id", storeProxy.Proxy)

	// Payment service routes
	paymentProxy := proxy.NewServiceProxy(cfg.Services.PaymentServiceURL)
	protected.Post("/payments", paymentProxy.Proxy)
	protected.Get("/payments/:id", paymentProxy.Proxy)

	// Inventory service routes
	inventoryProxy := proxy.NewServiceProxy(cfg.Services.InventoryServiceURL)
	protected.Get("/inventory", inventoryProxy.Proxy)
	protected.Get("/inventory/:id", inventoryProxy.Proxy)
	protected.Put("/inventory/:id", inventoryProxy.Proxy)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Infof("API Gateway listening on %s", addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	log.Info("API Gateway stopped")
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

func readinessCheck(c *fiber.Ctx) error {
	// Check dependencies (Redis, database, etc.)
	return c.JSON(fiber.Map{
		"status":    "ready",
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

	// Never expose stack traces to clients
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

// Placeholder handlers - to be implemented
func handleLogin(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Login endpoint - to be implemented"})
	}
}

func handleRefresh(jwtManager *auth.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Refresh endpoint - to be implemented"})
	}
}

func handleLogout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Logout endpoint - to be implemented"})
}
