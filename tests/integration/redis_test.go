package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/redis/go-redis/v9"
)

func TestRedisConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Start Redis container
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// Get connection string
	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	defer client.Close()

	// Test connection
	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	// Test set/get
	err = client.Set(ctx, "test-key", "test-value", time.Minute).Err()
	require.NoError(t, err)

	val, err := client.Get(ctx, "test-key").Result()
	require.NoError(t, err)
	require.Equal(t, "test-value", val)
}

