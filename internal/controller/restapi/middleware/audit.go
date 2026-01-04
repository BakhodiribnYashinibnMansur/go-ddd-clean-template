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
				m.logger.Errorw("failed to save endpoint history", zap.Error(err))
			}
		}(history)
	}
}
