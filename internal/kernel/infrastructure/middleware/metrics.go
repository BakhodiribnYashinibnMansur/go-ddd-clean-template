package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// OTelMetrics returns a Gin middleware that records HTTP request metrics using OpenTelemetry.
// Metrics: http_requests_total (counter), http_request_duration_seconds (histogram), http_requests_in_flight (gauge).
func OTelMetrics(serviceName string) gin.HandlerFunc {
	meter := otel.Meter(serviceName + "/http")

	requestCount, _ := meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	requestDuration, _ := meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	requestsInFlight, _ := meter.Int64UpDownCounter("http_requests_in_flight",
		metric.WithDescription("Number of HTTP requests currently being processed"),
	)

	return func(c *gin.Context) {
		requestsInFlight.Add(c.Request.Context(), 1)
		start := time.Now()

		c.Next()

		requestsInFlight.Add(c.Request.Context(), -1)

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		attrs := metric.WithAttributes(
			attribute.String("method", c.Request.Method),
			attribute.String("path", path),
			attribute.String("status", strconv.Itoa(c.Writer.Status())),
		)

		requestCount.Add(c.Request.Context(), 1, attrs)
		requestDuration.Record(c.Request.Context(), time.Since(start).Seconds(), attrs)
	}
}
