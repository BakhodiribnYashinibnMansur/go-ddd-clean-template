package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type healthDeps struct {
	pgPool *pgxpool.Pool
	redis  *redis.Client
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

		checks := make(map[string]string)
		allOK := true

		if deps.pgPool != nil {
			if err := deps.pgPool.Ping(ctx); err != nil {
				checks["postgres"] = "unhealthy: " + err.Error()
				allOK = false
			} else {
				checks["postgres"] = "ok"
			}
		} else {
			checks["postgres"] = "not configured"
		}

		if deps.redis != nil {
			if err := deps.redis.Ping(ctx).Err(); err != nil {
				checks["redis"] = "unhealthy: " + err.Error()
				allOK = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "not configured"
		}

		status := "ok"
		statusCode := http.StatusOK
		if !allOK {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status": status,
			"checks": checks,
			"uptime": time.Since(startTime).String(),
		})
	}
}
