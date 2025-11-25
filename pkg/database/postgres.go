package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/pkg/config"
	"github.com/onichange/pos-system/pkg/logger"
)

// PostgresDB wraps pgxpool.Pool with health check
type PostgresDB struct {
	Pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg config.DatabaseConfig, log *logger.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// Health check configuration
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Infof("Connected to PostgreSQL database: %s", cfg.DBName)

	return &PostgresDB{
		Pool:   pool,
		logger: log,
	}, nil
}

// HealthCheck checks database connection health
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	db.Pool.Close()
	db.logger.Info("PostgreSQL connection pool closed")
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
