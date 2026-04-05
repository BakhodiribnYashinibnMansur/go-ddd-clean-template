package middleware

import (
	"bytes"
	"io"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

const maxBodyLog = 4096 // max bytes to log from request body

// DebugBody logs request body at debug level for troubleshooting.
// Only active when the logger is at debug level.
func DebugBody(l logger.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		// Read body (up to maxBodyLog bytes)
		limit := c.Request.ContentLength
		if limit > maxBodyLog || limit < 0 {
			limit = maxBodyLog
		}

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, limit))
		if err != nil {
			c.Next()
			return
		}

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		// But we also need the full original body if it was truncated
		if c.Request.ContentLength > int64(len(body)) {
			remaining, _ := io.ReadAll(c.Request.Body)
			full := append(body, remaining...)
			c.Request.Body = io.NopCloser(bytes.NewReader(full))
		}

		ctx := c.Request.Context()
		l.Debugc(ctx, "request body",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"body", string(body),
			"content_length", c.Request.ContentLength,
		)

		c.Next()
	}
}
