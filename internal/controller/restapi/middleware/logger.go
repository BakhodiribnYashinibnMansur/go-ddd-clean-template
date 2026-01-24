package middleware

import (
	"time"

	"gct/internal/controller/restapi/util"
	"gct/pkg/logger"

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

		cp := c.Copy()
		requestID := c.GetString("request_id")

		go func() {
			l.Infow("HTTP request",
				zap.String("request_id", requestID),
				zap.String("method", cp.Request.Method),
				zap.String("path", path),
				zap.Int("status", cp.Writer.Status()),
				zap.String("latency", time.Since(start).String()),
				zap.String("client_ip", util.GetIPAddress(cp)),
				zap.String("error", cp.Errors.String()),
			)
		}()
	}
}
