// Package middleware contains Gin handlers for common API cross-cutting concerns.
package middleware

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/pkg/httpx"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditMiddleware provides functionality to record API usage and mutation history.
// It integrates with central auditing use cases to persist metrics and compliance logs.
type AuditMiddleware struct {
	uc     *usecase.UseCase
	logger logger.Log
}

// NewAuditMiddleware initializes an AuditMiddleware with access to core usecases and logging.
func NewAuditMiddleware(uc *usecase.UseCase, l logger.Log) *AuditMiddleware {
	return &AuditMiddleware{uc: uc, logger: l}
}

// EndpointHistory records high-level metrics (latency, status, path) for every processed request.
// It runs as a post-processing step after the main handler execution.
// This middleware helps in monitoring API usage patterns, performance bottlenecks, and debugging issues.
func (m *AuditMiddleware) EndpointHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture the start time to calculate request duration later.
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Parse the Request ID for cross-service tracing if provided in headers.
		// This ID helps correlate logs across multiple services in a distributed system.
		reqIDStr := httpx.GetCtxRequestID(c)
		var reqID *uuid.UUID
		if reqIDStr != "" {
			uid, err := uuid.Parse(reqIDStr)
			if err == nil {
				reqID = &uid
			}
		}

		// Proceed to actual endpoint logic.
		// All subsequent middlewares and the final handler will execute here.
		c.Next()

		// After the request is processed, collect execution metrics.
		duration := time.Since(start)
		ip := httpx.GetIPAddress(c)
		ua := httpx.GetUserAgent(c)
		respSize := c.Writer.Size()

		// Capture any errors that occurred during request processing.
		errMsg := ""
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		// Build the history record with all collected metrics.
		history := &domain.EndpointHistory{
			Method:       method,
			Path:         path,
			StatusCode:   c.Writer.Status(),
			DurationMs:   int(duration.Milliseconds()),
			IPAddress:    &ip,
			UserAgent:    &ua,
			CreatedAt:    time.Now(),
			RequestID:    reqID,
			ResponseSize: &respSize,
		}

		// Attach error message if any errors occurred.
		if errMsg != "" {
			history.ErrorMessage = &errMsg
		}

		// Link the identity and session if authentication was successful.
		// This allows tracking which user made which request.
		if session, err := httpx.GetCtxSession(c); err == nil {
			history.SessionID = &session.ID
			history.UserID = &session.UserID
		}

		// Persist the history entry asynchronously to avoid adding latency to the client response.
		// Uses a background context with timeout to ensure the goroutine outlives the request lifetime.
		go func(h *domain.EndpointHistory) {
			ctx, cancel := context.WithTimeout(context.Background(), consts.AuditPersistTimeout*time.Second)
			defer cancel()

			err := m.uc.Audit.History.Create(ctx, h)
			if err != nil {
				m.logger.Errorw("failed to save endpoint history", zap.Error(err))
			}
		}(history)
	}
}

// ChangeAudit explicitly records mutating actions (POST, PUT, DELETE, PATCH) to track administrative or sensitive state changes.
// This is crucial for compliance, security auditing, and forensic analysis.
// Read-only operations (GET, HEAD, OPTIONS) are skipped to reduce database load.
func (m *AuditMiddleware) ChangeAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Only mutation methods are subject to change auditing.
		// Read operations don't modify state, so we skip them.
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// Proceed to mutation logic.
		// The actual state change happens in the handler.
		c.Next()

		// Capture environmental and outcome data after the operation completes.
		ip := httpx.GetIPAddress(c)
		ua := httpx.GetUserAgent(c)

		// Build the audit log entry with operation metadata.
		auditLog := &domain.AuditLog{
			ID:        uuid.New(),
			Action:    domain.AuditActionAdminChange,
			IPAddress: &ip,
			UserAgent: &ua,
			Success:   c.Writer.Status() < consts.HTTPStatusSuccessThreshold, // HTTP status < 400 indicates success
			CreatedAt: time.Now(),
		}

		// Capture error details if the operation failed.
		if len(c.Errors) > 0 {
			errStr := c.Errors.String()
			auditLog.ErrorMessage = &errStr
		}

		// Identity linkage for accountability.
		// This ensures we know who performed the action.
		if session, err := httpx.GetCtxSession(c); err == nil {
			auditLog.SessionID = &session.ID
			auditLog.UserID = &session.UserID
		}

		// Comprehensive metadata helpful for security forensic analysis.
		// This includes the endpoint, method, status code, and query parameters.
		auditLog.Metadata = map[string]any{
			"path":   path,
			"method": method,
			"status": c.Writer.Status(),
			"query":  c.Request.URL.RawQuery,
		}

		// Asynchronously save mutation record to avoid blocking the response.
		go func(al *domain.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), consts.AuditPersistTimeout*time.Second)
			defer cancel()

			err := m.uc.Audit.Log.Create(ctx, al)
			if err != nil {
				m.logger.Errorw("failed to save change audit log", zap.Error(err))
			}
		}(auditLog)
	}
}
