// Package mongodb implements MongoDB connection.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"

	"gct/config"
	"gct/pkg/logger"
)

const (
	// ENV PROD
	maxPoolSizeProd = 50
	minPoolSizeProd = 10

	// ENV DEV
	maxPoolSizeDev = 8
	minPoolSizeDev = 3

	defaultConnectTimeout  = 10 * time.Second
	defaultPingTimeout     = 5 * time.Second
	defaultMaxConnIdleTime = 30 * time.Minute
)

// MongoDB struct wraps mongo.Client.
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// New creates a new MongoDB client with optimized settings.
func New(ctx context.Context, env string, cfg config.MongoDB, l logger.Log, opts ...Option) (*MongoDB, error) {
	uri := buildConnectionURI(cfg)

	clientOpts := options.Client().ApplyURI(uri)

	// Apply pool configuration
	applyPoolConfig(env, clientOpts)

	// Apply custom options
	for _, opt := range opts {
		opt(clientOpts)
	}

	// Create connection with timeout context
	connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		l.Errorw("failed to connect to MongoDB", zap.Error(err))
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, client, l); err != nil {
		if disconnectErr := client.Disconnect(ctx); disconnectErr != nil {
			l.Warnw("failed to disconnect MongoDB client during cleanup", zap.Error(disconnectErr))
		}
		return nil, err
	}

	m := &MongoDB{
		Client: client,
		DB:     client.Database(cfg.Name),
	}

	l.Infow("MongoDB connected successfully")

	return m, nil
}

// buildConnectionURI constructs the MongoDB connection URI.
func buildConnectionURI(cfg config.MongoDB) string {
	if cfg.User != "" && cfg.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	}
	return fmt.Sprintf("mongodb://%s:%d/%s", cfg.Host, cfg.Port, cfg.Name)
}

// applyPoolConfig applies pool configuration with defaults.
func applyPoolConfig(env string, clientOpts *options.ClientOptions) {
	var maxPoolSize, minPoolSize uint64

	if env == "production" || env == "PROD" {
		maxPoolSize = maxPoolSizeProd
		minPoolSize = minPoolSizeProd
	} else {
		maxPoolSize = maxPoolSizeDev
		minPoolSize = minPoolSizeDev
	}

	clientOpts.SetMaxPoolSize(maxPoolSize)
	clientOpts.SetMinPoolSize(minPoolSize)
	clientOpts.SetMaxConnIdleTime(defaultMaxConnIdleTime)
}

// verifyConnection pings the MongoDB server to ensure connectivity.
func verifyConnection(ctx context.Context, client *mongo.Client, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		l.Errorw("failed to ping MongoDB server", zap.Error(err))
		return fmt.Errorf("verify mongodb connection: %w", err)
	}

	return nil
}

// Close gracefully closes the MongoDB connection.
func (m *MongoDB) Close(ctx context.Context) error {
	if m != nil && m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}
