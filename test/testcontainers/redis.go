package container

import (
	"context"
	"fmt"
	"log"

	"gct/config"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunRedisTestContainer is a function that runs a Redis test container
// RunRedisTestContainer runs a Redis test container
func RunRedisTestContainer(cfg config.RedisStore) (*redis.Client, testcontainers.Container, error) {
	// Test-infrastructure bootstrap — this function is only used to spin up
	// containers during test suites / local setup; no caller context applies.
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        RedisImage,
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithOccurrence(1),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	// Get Redis host and port
	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, redisContainer, fmt.Errorf("failed to get redis host: %w", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, redisContainer, fmt.Errorf("failed to get redis port: %w", err)
	}

	redisAddr := fmt.Sprintf("%s:%s", host, port.Port())
	log.Printf("Redis address: %s", redisAddr)

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, redisContainer, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Printf("Redis test container ready")

	return rdb, redisContainer, nil
}
