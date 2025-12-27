package middleware

import (
	"time"

	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger -.
func Logger(l logger.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		l.Infow("HTTP request", zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("latency", time.Since(start).String()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("error", c.Errors.String()),
		)
	}
}
