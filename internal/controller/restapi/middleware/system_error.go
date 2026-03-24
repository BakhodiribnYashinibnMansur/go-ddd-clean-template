package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SystemErrorMiddleware captures critical failures (panics and 5xx errors) and persists them
// to the database for audit and diagnostic purposes.
type SystemErrorMiddleware struct {
	uc     *usecase.UseCase
	logger logger.Log
}

// NewSystemErrorMiddleware initializes the error tracking middleware.
func NewSystemErrorMiddleware(uc *usecase.UseCase, l logger.Log) *SystemErrorMiddleware {
	return &SystemErrorMiddleware{uc: uc, logger: l}
}

// Recovery serves as a specialized panic handler.
// It intercepts runtime panics, logs the stack trace, and saves a "PANIC" record to the system_errors table.
// It ensures that the API returns a clean 500 JSON response instead of crashing the connection.
func (m *SystemErrorMiddleware) Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		start := time.Now()
		stack := string(debug.Stack()) // Capture stack trace for debugging

		m.logger.Errorw("panic recovered",
			"error", recovered,
			"stack", stack,
		)

		// Persist the panic event to the database.
		m.saveError(c, recovered, &stack, "PANIC", start)

		// Return standardized error response.
		response.ControllerResponse(c, http.StatusInternalServerError, httpx.ErrPanicRecovered, nil, false)
		c.Abort()
	})
}

// Persist5xx inspects the response status and error list after the handler chain finishes.
// If any 500-level error occurred during safe execution (non-panic), it is logged to the system_errors table.
func (m *SystemErrorMiddleware) Persist5xx() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // Wait for downstream handlers to complete

		// Check if any errors were appended to the context OR if the status code is server error.
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			// Iterate through all accumulated errors and persist them.
			for _, e := range c.Errors {
				m.saveError(c, e.Err, nil, "ERROR", start)
			}

			// If no explicit error object was attached but a 500 was returned, log a generic event.
			if len(c.Errors) == 0 && c.Writer.Status() >= 500 {
				m.saveError(c, "Unknown internal server error (No error object attached)", nil, "ERROR", start)
			}
		}
	}
}

// saveError constructs the SystemError domain object and dispatches it to the repository asynchronously.
func (m *SystemErrorMiddleware) saveError(c *gin.Context, errVal any, stack *string, severity string, start time.Time) {
	// Extract request context for metadata
	path := c.Request.URL.Path
	method := c.Request.Method
	ip := httpx.GetIPAddress(c)

	// Build the error record
	sysErr := &domain.SystemError{
		Code:        "INTERNAL_ERROR",
		Message:     fmt.Sprintf("%v", errVal),
		StackTrace:  stack,
		Severity:    severity,
		ServiceName: stringPtr("api"),
		Path:        &path,
		Method:      &method,
		IPAddress:   &ip,
		CreatedAt:   time.Now(),
		Metadata: map[string]any{
			"duration_ms": time.Since(start).Milliseconds(),
			"user_agent":  httpx.GetUserAgent(c),
		},
	}

	// Link Request ID if available (from RequestID middleware)
	reqIDStr := httpx.GetCtxRequestID(c)
	if reqIDStr != "" {
		if uid, err := uuid.Parse(reqIDStr); err == nil {
			sysErr.RequestID = &uid
		}
	}

	// Link User ID if authenticated (from Auth middleware)
	userIDStr := c.GetString(consts.CtxUserID)
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			sysErr.UserID = &uid
		}
	}

	// Fire-and-forget: Persist to DB in background context
	// Uses a disconnected context with a timeout to ensure persistence happens even if client disconnects.
	go func(val *domain.SystemError) {
		ctx, cancel := context.WithTimeout(context.Background(), consts.DurationAuditSave*time.Second)
		defer cancel()

		if err := m.uc.Audit.SystemError().Create(ctx, val); err != nil {
			m.logger.Errorw("failed to persist system error", zap.Error(err), "original_error", val.Message)
		}
	}(sysErr)
}

// Helper for inline pointer creation.
func stringPtr(s string) *string {
	return &s
}
