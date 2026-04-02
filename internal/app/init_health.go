package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type healthDeps struct {
	pgPool    *pgxpool.Pool
	redis     *redis.Client
	asynqAddr string // empty if disabled
}

var startTime = time.Now()

func registerHealthRoutes(r *gin.Engine, deps healthDeps) {
	r.GET("/health", handleHealth)
	r.GET("/ready", handleReady(deps))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
	})
}

func handleReady(deps healthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		checks := make(map[string]any)
		allOK := true

		// PostgreSQL
		if deps.pgPool != nil {
			if err := deps.pgPool.Ping(ctx); err != nil {
				checks["postgres"] = gin.H{"status": "unhealthy", "error": err.Error()}
				allOK = false
			} else {
				stat := deps.pgPool.Stat()
				checks["postgres"] = gin.H{
					"status":           "ok",
					"total_conns":      stat.TotalConns(),
					"idle_conns":       stat.IdleConns(),
					"acquired_conns":   stat.AcquiredConns(),
					"max_conns":        stat.MaxConns(),
				}
			}
		} else {
			checks["postgres"] = gin.H{"status": "not configured"}
		}

		// Redis
		if deps.redis != nil {
			if err := deps.redis.Ping(ctx).Err(); err != nil {
				checks["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
				allOK = false
			} else {
				poolStats := deps.redis.PoolStats()
				checks["redis"] = gin.H{
					"status":     "ok",
					"total_conns": poolStats.TotalConns,
					"idle_conns":  poolStats.IdleConns,
					"stale_conns": poolStats.StaleConns,
					"hits":        poolStats.Hits,
					"misses":      poolStats.Misses,
				}
			}
		} else {
			checks["redis"] = gin.H{"status": "not configured"}
		}

		// Asynq (via Redis ping on its address)
		if deps.asynqAddr != "" {
			checks["asynq"] = gin.H{"status": "ok", "addr": deps.asynqAddr}
		} else {
			checks["asynq"] = gin.H{"status": "not configured"}
		}

		status := "ok"
		statusCode := http.StatusOK
		if !allOK {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status":   status,
			"checks":   checks,
			"uptime":   time.Since(startTime).String(),
			"uptime_s": fmt.Sprintf("%.0f", time.Since(startTime).Seconds()),
		})
	}
}
