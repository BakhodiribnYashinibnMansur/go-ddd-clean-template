package middleware

import (
	"context"

	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID middleware generates a unique ID for the request if not already present.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}

		c.Header("X-Request-ID", id)

		// Set in standard context
		ctx := context.WithValue(c.Request.Context(), logger.KeyRequestID, id)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
