package middleware

import (
	"context"

	"gct/consts"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID returns a Gin middleware that ensures every HTTP request has a unique identifier.
// If the client provides an "X-Request-ID" header, it is propagated (useful for service-to-service calls);
// otherwise, a new UUID v4 is generated.
//
// This ID is used for log correlation and cross-service tracing throughout the request lifecycle.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(consts.HeaderXRequestID)
		if id == "" {
			// Generate fresh ID for internal traceability
			id = uuid.New().String()
		}

		// Ensure the client receives the ID back in the response headers.
		// Allowed for frontend debugging.
		c.Header(consts.HeaderXRequestID, id)

		// Persist the ID in the Go standard context so it can be picked up by logger packages
		// and deeper layers of the application (usecase, repository) without coupling to Gin.
		ctx := context.WithValue(c.Request.Context(), logger.KeyRequestID, id)
		c.Request = c.Request.WithContext(ctx)

		// Map to Gin context for easy retrieval in controllers.
		c.Set(consts.CtxKeyRequestID, id)

		c.Next()
	}
}
