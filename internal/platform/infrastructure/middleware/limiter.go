package middleware

import (
	"fmt"
	"gct/internal/platform/infrastructure/httpx"
	"net/http"

	"gct/config"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
	libredis "github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.uber.org/zap"
)

// RateLimiter returns a Gin middleware that enforces request limits using a sliding window algorithm.
// It uses Redis as a centralized store to support rate limiting across multiple application instances (distributed rate limiting).
func RateLimiter(cfg config.Limiter, client *libredis.Client, l logger.Log) gin.HandlerFunc {
	// Bypass if rate limiting is disabled in configuration.
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Define the allowed rate (e.g., 100 requests per minute).
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", cfg.Limit, cfg.Period))
	if err != nil {
		l.Errorw("failed to parse rate limit", zap.Error(err))
		// Fallback to a safe production default of 100 requests/minute if config parsing fails.
		rate = limiter.Rate{
			Period: 60,
			Limit:  100,
		}
	}

	// Initialize the Redis-backed store for persisting hit counts.
	// The prefix ensures no collision with other Redis keys.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: "limiter:api:",
	})
	if err != nil {
		l.Errorw("failed to create limiter store", zap.Error(err))
		// Fail-open strategy: allow traffic if the limiter store (Redis) is unreachable.
		// This prevents the rate limiter from causing a system-wide outage during cache issues.
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Instantiate the core limiter engine.
	instance := limiter.New(store, rate)

	// Construct the Gin-specific middleware adapter.
	return mgin.NewMiddleware(instance, mgin.WithErrorHandler(func(c *gin.Context, err error) {
		// Log internal failures in the limiter logic, but do not block user (Fail Open).
		l.Warnw("rate limiter error - failing open", zap.String("error", err.Error()))
		c.Next()
	}), mgin.WithLimitReachedHandler(func(c *gin.Context) {
		// Handle cases where the client exceeds their allowed quota.
		l.Warnw("rate limit reached",
			zap.String("ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)
		response.ControllerResponse(c, http.StatusTooManyRequests, httpx.ErrRateLimitExceeded, nil, false)
		c.Abort()
	}))
}
