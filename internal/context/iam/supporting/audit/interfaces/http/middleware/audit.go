// Package middleware contains Gin handlers for audit cross-cutting concerns.
package middleware

import (
	"context"
	"fmt"
	"time"

	"gct/internal/context/iam/supporting/audit/application/command"
	auditdomain "gct/internal/context/iam/supporting/audit/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditMiddleware provides functionality to record API usage and mutation history.
// It uses DDD command handlers from the Audit bounded context to persist metrics and compliance logs.
type AuditMiddleware struct {
	createEndpointHistory *command.CreateEndpointHistoryHandler
	createAuditLog        *command.CreateAuditLogHandler
	l                     logger.Log
}

// NewAuditMiddleware initializes an AuditMiddleware with the required DDD command handlers.
func NewAuditMiddleware(
	createHistory *command.CreateEndpointHistoryHandler,
	createLog *command.CreateAuditLogHandler,
	l logger.Log,
) *AuditMiddleware {
	return &AuditMiddleware{
		createEndpointHistory: createHistory,
		createAuditLog:        createLog,
		l:                     l,
	}
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

		// Proceed to actual endpoint logic.
		// All subsequent middlewares and the final handler will execute here.
		c.Next()

		// After the request is processed, collect execution metrics.
		duration := time.Since(start)
		ip := httpx.GetIPAddress(c)
		ua := httpx.GetUserAgent(c)

		// Build the command with all collected metrics.
		cmd := command.CreateEndpointHistoryCommand{
			Endpoint:   path,
			Method:     method,
			StatusCode: c.Writer.Status(),
			Latency:    int(duration.Milliseconds()),
			IPAddress:  &ip,
			UserAgent:  &ua,
		}

		// Link the identity if authentication was successful.
		// This allows tracking which user made which request.
		if sessionVal, exists := c.Get(consts.CtxSession); exists {
			if session, ok := sessionVal.(*shared.AuthSession); ok {
				cmd.UserID = &session.UserID
			}
		}

		// Persist the history entry asynchronously to avoid adding latency to the client response.
		// Detach cancellation from the request context so the goroutine outlives the response,
		// while still preserving values (trace IDs, logger fields) from the parent context.
		bgCtx := context.WithoutCancel(c.Request.Context())
		go func(cmd command.CreateEndpointHistoryCommand) {
			ctx, cancel := context.WithTimeout(bgCtx, consts.AuditPersistTimeout*time.Second)
			defer cancel()

			if err := m.createEndpointHistory.Handle(ctx, cmd); err != nil {
				m.l.Errorw("failed to save endpoint history", zap.Error(err))
			}
		}(cmd)
	}
}

// ChangeAudit explicitly records mutating actions (POST, PUT, DELETE, PATCH) to track
// administrative or sensitive state changes.
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

		// Build the audit log command with operation metadata.
		cmd := command.CreateAuditLogCommand{
			Action:    auditdomain.AuditActionAdminChange,
			IPAddress: &ip,
			UserAgent: &ua,
			Success:   c.Writer.Status() < consts.HTTPStatusSuccessThreshold, // HTTP status < 400 indicates success
			Metadata: map[string]string{
				"path":   path,
				"method": method,
				"status": fmt.Sprintf("%d", c.Writer.Status()),
				"query":  c.Request.URL.RawQuery,
			},
		}

		// Capture error details if the operation failed.
		if len(c.Errors) > 0 {
			errStr := c.Errors.String()
			cmd.ErrorMessage = &errStr
		}

		// Identity linkage for accountability.
		// This ensures we know who performed the action.
		if sessionVal, exists := c.Get(consts.CtxSession); exists {
			if session, ok := sessionVal.(*shared.AuthSession); ok {
				cmd.SessionID = &session.ID
				cmd.UserID = &session.UserID
			}
		}

		// Asynchronously save mutation record to avoid blocking the response.
		// Detach cancellation from the request context so the goroutine outlives the response,
		// while still preserving values from the parent context.
		bgCtx := context.WithoutCancel(c.Request.Context())
		go func(cmd command.CreateAuditLogCommand) {
			ctx, cancel := context.WithTimeout(bgCtx, consts.AuditPersistTimeout*time.Second)
			defer cancel()

			if err := m.createAuditLog.Handle(ctx, cmd); err != nil {
				m.l.Errorw("failed to save change audit log", zap.Error(err))
			}
		}(cmd)
	}
}

// getSessionFromContext is a helper to safely extract the AuthSession from the Gin context.
func getSessionFromContext(c *gin.Context) (*shared.AuthSession, bool) {
	sessionVal, exists := c.Get(consts.CtxSession)
	if !exists {
		return nil, false
	}
	session, ok := sessionVal.(*shared.AuthSession)
	return session, ok
}

// uuidPtr returns a pointer to a new UUID parsed from the given string, or nil on failure.
func uuidPtr(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	uid, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &uid
}
