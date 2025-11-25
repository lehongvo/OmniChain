package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Security SecurityConfig
	Services ServicesConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Environment  string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MinConnections  int
	MaxConnLifetime time.Duration
	ConnMaxIdleTime time.Duration
	QueryTimeout    time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	MinIdleConns int
	PoolSize     int
	PoolTimeout  time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	RateLimitRequestsPerMinute int
	RateLimitBurst             int
	MaxRequestSize             int64
	EnableCORS                 bool
	CORSOrigins                []string
	EnableTLS                  bool
	TLSCertPath                string
	TLSKeyPath                 string
}

// ServicesConfig holds microservices configuration
type ServicesConfig struct {
	OrderServiceURL        string
	UserServiceURL         string
	StoreServiceURL        string
	PaymentServiceURL      string
	InventoryServiceURL    string
	NotificationServiceURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "onichange"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConnections:  getIntEnv("DB_MAX_CONNECTIONS", 100),
			MinConnections:  getIntEnv("DB_MIN_CONNECTIONS", 10),
			MaxConnLifetime: getDurationEnv("DB_MAX_CONN_LIFETIME", 1*time.Hour),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 30*time.Minute),
			QueryTimeout:    getDurationEnv("DB_QUERY_TIMEOUT", 30*time.Second),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 20),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 200),
			PoolTimeout:  getDurationEnv("REDIS_POOL_TIMEOUT", 5*time.Second),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTokenExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			Issuer:             getEnv("JWT_ISSUER", "onichange"),
		},
		Security: SecurityConfig{
			RateLimitRequestsPerMinute: getIntEnv("RATE_LIMIT_REQUESTS", 100),
			RateLimitBurst:             getIntEnv("RATE_LIMIT_BURST", 10),
			MaxRequestSize:             getInt64Env("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
			EnableCORS:                 getBoolEnv("ENABLE_CORS", true),
			CORSOrigins:                getStringSliceEnv("CORS_ORIGINS", []string{"*"}),
			EnableTLS:                  getBoolEnv("ENABLE_TLS", false),
			TLSCertPath:                getEnv("TLS_CERT_PATH", ""),
			TLSKeyPath:                 getEnv("TLS_KEY_PATH", ""),
		},
		Services: ServicesConfig{
			OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
			UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8082"),
			StoreServiceURL:        getEnv("STORE_SERVICE_URL", "http://localhost:8083"),
			PaymentServiceURL:      getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084"),
			InventoryServiceURL:    getEnv("INVENTORY_SERVICE_URL", "http://localhost:8085"),
			NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8086"),
		},
	}

	// Validate required fields
	if config.JWT.AccessTokenSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if config.JWT.RefreshTokenSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
	}

	return config, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In production, use a proper CSV parser
		return []string{value}
	}
	return defaultValue
}
