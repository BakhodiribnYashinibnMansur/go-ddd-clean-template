package middleware

import (
	"time"

	"gct/internal/shared/infrastructure/metrics/latency"

	"github.com/gin-gonic/gin"
)

// LatencyTracker returns a Gin middleware that records request duration
// into the provided latency tracker.
func LatencyTracker(tracker *latency.Tracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		tracker.Record(time.Since(start))
	}
}
