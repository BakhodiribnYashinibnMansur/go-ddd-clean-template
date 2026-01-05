package middleware

import (
	"fmt"
	"net/http"

	"gct/config"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	libredis "github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.uber.org/zap"
)

// RateLimiter returns a gin middleware for rate limiting.
func RateLimiter(cfg config.Limiter, client *libredis.Client, l logger.Log) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Define a rate.
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%d-%s", cfg.Limit, cfg.Period))
	if err != nil {
		l.Errorw("failed to parse rate limit", zap.Error(err))
		// Fallback to a safe default if parsing fails
		rate = limiter.Rate{
			Period: 60,
			Limit:  100,
		}
	}

	// Create a redis store.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: "limiter:api:",
	})
	if err != nil {
		l.Errorw("failed to create limiter store", zap.Error(err))
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Create a new limiter instance.
	instance := limiter.New(store, rate)

	// Create a new middleware instance.
	return mgin.NewMiddleware(instance, mgin.WithErrorHandler(func(c *gin.Context, err error) {
		l.WithContext(c.Request.Context()).Errorw("rate limiter error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		c.Abort()
	}), mgin.WithLimitReachedHandler(func(c *gin.Context) {
		l.WithContext(c.Request.Context()).Warnw("rate limit reached",
			zap.String("ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many requests",
		})
		c.Abort()
	}))
}
