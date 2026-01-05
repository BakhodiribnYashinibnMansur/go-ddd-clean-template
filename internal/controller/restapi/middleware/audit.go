package middleware

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuditMiddleware struct {
	uc     *usecase.UseCase
	logger logger.Log
}

func NewAuditMiddleware(uc *usecase.UseCase, l logger.Log) *AuditMiddleware {
	return &AuditMiddleware{uc: uc, logger: l}
}

func (m *AuditMiddleware) EndpointHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		// Request ID
		reqIDStr := util.GetRequestID(c)
		var reqID *uuid.UUID
		if reqIDStr != "" {
			uid, err := uuid.Parse(reqIDStr)
			if err == nil {
				reqID = &uid
			}
		}

		// Process request
		c.Next()

		duration := time.Since(start)

		ip := util.GetIPAddress(c)
		ua := util.GetUserAgent(c)
		respSize := c.Writer.Size()
		errMsg := ""
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

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

		if errMsg != "" {
			history.ErrorMessage = &errMsg
		}

		// Extract user/session if available
		if sessionVal, exists := c.Get(consts.CtxSession); exists {
			if session, ok := sessionVal.(*domain.Session); ok {
				history.SessionID = &session.ID
				history.UserID = &session.UserID
			}
		}

		// Async save to avoid blocking response
		// Note: Creating a detached context or using background
		go func(h *domain.EndpointHistory) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// We can't use c.Request.Context() because it might be cancelled
			// Use background context

			err := m.uc.Audit.History.Create(ctx, h)
			if err != nil {
				// Use zap type as requested
				m.logger.WithContext(ctx).Errorw("failed to save endpoint history", zap.Error(err))
			}
		}(history)
	}
}

// ChangeAudit records mutating actions (POST, PUT, DELETE, PATCH) to the audit log.
func (m *AuditMiddleware) ChangeAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// We only care about mutations
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Record if successful or if it's an admin path
		// Typically we log all mutations in admin panel, even if they failed (for security auditing)

		ip := util.GetIPAddress(c)
		ua := util.GetUserAgent(c)

		auditLog := &domain.AuditLog{
			ID:        uuid.New(),
			Action:    domain.AuditActionAdminChange,
			IPAddress: &ip,
			UserAgent: &ua,
			Success:   c.Writer.Status() < 400,
			CreatedAt: time.Now(),
		}

		if len(c.Errors) > 0 {
			errStr := c.Errors.String()
			auditLog.ErrorMessage = &errStr
		}

		// Extract user/session if available
		if sessionVal, exists := c.Get(consts.CtxSession); exists {
			if session, ok := sessionVal.(*domain.Session); ok {
				auditLog.SessionID = &session.ID
				auditLog.UserID = &session.UserID
			}
		}

		// Metadata: capture detailed info for "Who did what"
		auditLog.Metadata = map[string]any{
			"path":   path,
			"method": method,
			"status": c.Writer.Status(),
			"query":  c.Request.URL.RawQuery,
		}

		// Async save
		go func(al *domain.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := m.uc.Audit.Log.Create(ctx, al)
			if err != nil {
				m.logger.WithContext(ctx).Errorw("failed to save change audit log", zap.Error(err))
			}
		}(auditLog)
	}
}
