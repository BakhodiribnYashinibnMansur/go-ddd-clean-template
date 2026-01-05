// Package featureflag provides feature flag functionality using GoFeatureFlag.
package featureflag

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/pkg/logger"

	"github.com/redis/go-redis/v9"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"go.uber.org/zap"
)

// Client wraps the GoFeatureFlag client.
type Client struct {
	logger logger.Log
}

// New creates a new feature flag client.
func New(ctx context.Context, cfg config.FeatureFlag, redisClient *redis.Client, l logger.Log) (*Client, error) {
	if !cfg.Enabled {
		l.WithContext(ctx).Infow("feature flags disabled")
		return &Client{logger: l}, nil
	}

	var retrievers []retriever.Retriever

	// File retriever
	if cfg.UseFileRetriever {
		fileRetriever := &fileretriever.Retriever{
			Path: cfg.ConfigPath,
		}
		retrievers = append(retrievers, fileRetriever)
		l.WithContext(ctx).Infow("feature flag file retriever configured",
			zap.String("path", cfg.ConfigPath))
	}

	// Redis retriever (if enabled)
	if cfg.UseRedis && redisClient != nil {
		redisRetriever := NewRedisRetriever(redisClient, cfg.RedisKey)
		retrievers = append(retrievers, redisRetriever)
		l.WithContext(ctx).Infow("feature flag redis retriever configured",
			zap.String("key", cfg.RedisKey))
	}

	if len(retrievers) == 0 {
		return nil, fmt.Errorf("no feature flag retrievers configured")
	}

	// Initialize GoFeatureFlag
	err := ffclient.Init(ffclient.Config{
		PollingInterval: time.Duration(cfg.PollingInterval) * time.Second,
		Retrievers:      retrievers,
		Context:         ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag client: %w", err)
	}

	l.WithContext(ctx).Infow("feature flag client initialized successfully",
		zap.Int("polling_interval", cfg.PollingInterval))

	return &Client{
		logger: l,
	}, nil
}

// Close closes the feature flag client.
func (c *Client) Close(ctx context.Context) {
	ffclient.Close()
	c.logger.WithContext(ctx).Infow("feature flag client closed")
}

// IsEnabled checks if a feature flag is enabled for a given user.
func (c *Client) IsEnabled(ctx context.Context, flagKey string, user User, defaultValue bool) bool {
	result, err := ffclient.BoolVariation(flagKey, user.ToEvaluationContext(), defaultValue)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("failed to evaluate feature flag",
			zap.String("flag", flagKey),
			zap.Error(err),
			zap.Bool("default", defaultValue))
		return defaultValue
	}

	c.logger.WithContext(ctx).Debugw("feature flag evaluated",
		zap.String("flag", flagKey),
		zap.Bool("result", result),
		zap.String("user_id", user.Key))

	return result
}

// GetStringVariation returns a string variation of a feature flag.
func (c *Client) GetStringVariation(ctx context.Context, flagKey string, user User, defaultValue string) string {
	result, err := ffclient.StringVariation(flagKey, user.ToEvaluationContext(), defaultValue)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("failed to evaluate feature flag",
			zap.String("flag", flagKey),
			zap.Error(err),
			zap.String("default", defaultValue))
		return defaultValue
	}

	c.logger.WithContext(ctx).Debugw("feature flag evaluated",
		zap.String("flag", flagKey),
		zap.String("result", result),
		zap.String("user_id", user.Key))

	return result
}

// GetIntVariation returns an int variation of a feature flag.
func (c *Client) GetIntVariation(ctx context.Context, flagKey string, user User, defaultValue int) int {
	result, err := ffclient.IntVariation(flagKey, user.ToEvaluationContext(), defaultValue)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("failed to evaluate feature flag",
			zap.String("flag", flagKey),
			zap.Error(err),
			zap.Int("default", defaultValue))
		return defaultValue
	}

	c.logger.WithContext(ctx).Debugw("feature flag evaluated",
		zap.String("flag", flagKey),
		zap.Int("result", result),
		zap.String("user_id", user.Key))

	return result
}

// GetFloatVariation returns a float variation of a feature flag.
func (c *Client) GetFloatVariation(ctx context.Context, flagKey string, user User, defaultValue float64) float64 {
	result, err := ffclient.Float64Variation(flagKey, user.ToEvaluationContext(), defaultValue)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("failed to evaluate feature flag",
			zap.String("flag", flagKey),
			zap.Error(err),
			zap.Float64("default", defaultValue))
		return defaultValue
	}

	c.logger.WithContext(ctx).Debugw("feature flag evaluated",
		zap.String("flag", flagKey),
		zap.Float64("result", result),
		zap.String("user_id", user.Key))

	return result
}

// GetJSONVariation returns a JSON variation of a feature flag.
func (c *Client) GetJSONVariation(ctx context.Context, flagKey string, user User, defaultValue map[string]interface{}) map[string]interface{} {
	result, err := ffclient.JSONVariation(flagKey, user.ToEvaluationContext(), defaultValue)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("failed to evaluate feature flag",
			zap.String("flag", flagKey),
			zap.Error(err))
		return defaultValue
	}

	c.logger.WithContext(ctx).Debugw("feature flag evaluated",
		zap.String("flag", flagKey),
		zap.String("user_id", user.Key))

	return result
}
