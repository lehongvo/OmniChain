package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/pkg/config"
	"github.com/onichange/pos-system/pkg/logger"
)

// ReplicaPool manages read replicas with load balancing
type ReplicaPool struct {
	primary  *pgxpool.Pool
	replicas []*pgxpool.Pool
	current  int
	logger   *logger.Logger
}

// NewReplicaPool creates a new replica pool
func NewReplicaPool(primaryConfig config.DatabaseConfig, replicaConfigs []config.DatabaseConfig, log *logger.Logger) (*ReplicaPool, error) {
	// Create primary connection
	primary, err := NewPostgresDB(primaryConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary connection: %w", err)
	}

	// Create replica connections
	replicas := make([]*pgxpool.Pool, 0, len(replicaConfigs))
	for _, cfg := range replicaConfigs {
		replica, err := NewPostgresDB(cfg, log)
		if err != nil {
			log.Warnf("Failed to create replica connection: %v", err)
			continue
		}
		replicas = append(replicas, replica.Pool)
	}

	return &ReplicaPool{
		primary:  primary.Pool,
		replicas: replicas,
		logger:   log,
	}, nil
}

// GetPrimary returns the primary database connection
func (rp *ReplicaPool) GetPrimary() *pgxpool.Pool {
	return rp.primary
}

// GetReplica returns a read replica (round-robin)
func (rp *ReplicaPool) GetReplica() *pgxpool.Pool {
	if len(rp.replicas) == 0 {
		return rp.primary // Fallback to primary if no replicas
	}

	rp.current = (rp.current + 1) % len(rp.replicas)
	return rp.replicas[rp.current]
}

// HealthCheck checks health of all connections
func (rp *ReplicaPool) HealthCheck(ctx context.Context) error {
	// Check primary
	if err := rp.primary.Ping(ctx); err != nil {
		return fmt.Errorf("primary connection unhealthy: %w", err)
	}

	// Check replicas
	for i, replica := range rp.replicas {
		if err := replica.Ping(ctx); err != nil {
			rp.logger.Warnf("Replica %d is unhealthy: %v", i, err)
		}
	}

	return nil
}

// Close closes all connections
func (rp *ReplicaPool) Close() {
	rp.primary.Close()
	for _, replica := range rp.replicas {
		replica.Close()
	}
}
