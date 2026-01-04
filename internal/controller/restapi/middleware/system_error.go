package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SystemErrorMiddleware struct {
	uc     *usecase.UseCase
	logger logger.Log
}

func NewSystemErrorMiddleware(uc *usecase.UseCase, l logger.Log) *SystemErrorMiddleware {
	return &SystemErrorMiddleware{uc: uc, logger: l}
}

func (m *SystemErrorMiddleware) Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		start := time.Now()
		stack := string(debug.Stack())

		m.logger.Errorw("panic recovered",
			"error", recovered,
			"stack", stack,
		)

		// Save to DB
		m.saveError(c, recovered, &stack, "PANIC", start)

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func (m *SystemErrorMiddleware) Persist5xx() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			// Skip if it was already handled by Recovery (which aborts, but Next() returns)
			// Actually Recovery intercepts panic, so this might run after recovery if configured correctly?
			// No, Recovery is usually outside/before.
			// But if Recovery handles it, it writes response.
			// Let's check status.

			// We iterate errors
			for _, e := range c.Errors {
				m.saveError(c, e.Err, nil, "ERROR", start)
			}

			// If no errors attached but status is 500, save generic
			if len(c.Errors) == 0 && c.Writer.Status() >= 500 {
				m.saveError(c, "Unknown internal server error", nil, "ERROR", start)
			}
		}
	}
}

func (m *SystemErrorMiddleware) saveError(c *gin.Context, errVal any, stack *string, severity string, start time.Time) {
	// Extract context info
	path := c.Request.URL.Path
	method := c.Request.Method
	ip := util.GetIPAddress(c)

	// Create struct
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
			"user_agent":  util.GetUserAgent(c),
		},
	}

	// Extract Request ID
	reqIDStr := util.GetRequestID(c)
	if reqIDStr != "" {
		if uid, err := uuid.Parse(reqIDStr); err == nil {
			sysErr.RequestID = &uid
		}
	}

	// Extract User ID
	userIDStr := c.GetString("user_id") // set by AuthMiddleware
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			sysErr.UserID = &uid
		}
	}

	// Async save
	go func(val *domain.SystemError) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.uc.Audit.SystemError.Create(ctx, val); err != nil {
			m.logger.Errorw("failed to persist system error", zap.Error(err), "original_error", val.Message)
		}
	}(sysErr)
}

func stringPtr(s string) *string {
	return &s
}
