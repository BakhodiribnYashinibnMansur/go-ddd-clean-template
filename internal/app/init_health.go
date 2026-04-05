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
	// K8s-style health endpoints:
	//   /health/live  — liveness: 200 if process is up, no dependency checks.
	//                   K8s restarts the pod if this fails.
	//   /health/ready — readiness: checks postgres/redis/asynq. K8s removes
	//                   the pod from the load balancer if this fails.
	//   /health       — backward-compat alias for liveness.
	r.GET("/health", handleLive)
	r.GET("/health/live", handleLive)
	r.GET("/health/ready", handleReady(deps))
}

func handleLive(c *gin.Context) {
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

		allOK = checkPostgres(ctx, deps, checks) && allOK
		allOK = checkRedis(ctx, deps, checks) && allOK
		checks["asynq"] = checkAsynq(deps)

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

// checkPostgres probes the Postgres pool and records its status in checks.
// Returns true if the component is healthy or not configured.
func checkPostgres(ctx context.Context, deps healthDeps, checks map[string]any) bool {
	if deps.pgPool == nil {
		checks["postgres"] = gin.H{"status": "not configured"}
		return true
	}
	if err := deps.pgPool.Ping(ctx); err != nil {
		checks["postgres"] = gin.H{"status": "unhealthy", "error": err.Error()}
		return false
	}
	stat := deps.pgPool.Stat()
	checks["postgres"] = gin.H{
		"status":         "ok",
		"total_conns":    stat.TotalConns(),
		"idle_conns":     stat.IdleConns(),
		"acquired_conns": stat.AcquiredConns(),
		"max_conns":      stat.MaxConns(),
	}
	return true
}

// checkRedis probes the Redis client and records its status in checks.
// Returns true if the component is healthy or not configured.
func checkRedis(ctx context.Context, deps healthDeps, checks map[string]any) bool {
	if deps.redis == nil {
		checks["redis"] = gin.H{"status": "not configured"}
		return true
	}
	if err := deps.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
		return false
	}
	poolStats := deps.redis.PoolStats()
	checks["redis"] = gin.H{
		"status":      "ok",
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
	}
	return true
}

// checkAsynq reports the configured Asynq address.
func checkAsynq(deps healthDeps) gin.H {
	if deps.asynqAddr == "" {
		return gin.H{"status": "not configured"}
	}
	return gin.H{"status": "ok", "addr": deps.asynqAddr}
}
