package errorx

import (
	"context"
	"fmt"

	"gct/internal/platform/domain/consts"
	"gct/internal/platform/infrastructure/logger"

	"github.com/google/uuid"
)

// ServiceErrorLogger provides service-specific error logging
type ServiceErrorLogger struct {
	errorLogger *ErrorLogger
	logger      logger.Log
	serviceName string
}

// NewServiceErrorLogger creates a new service error logger
func NewServiceErrorLogger(repo Repository, logger logger.Log, serviceName string) *ServiceErrorLogger {
	return &ServiceErrorLogger{
		errorLogger: NewErrorLogger(repo, logger),
		logger:      logger,
		serviceName: serviceName,
	}
}

// LogError logs a service error
func (s *ServiceErrorLogger) LogError(ctx context.Context, code string, message string, err error, metadata map[string]any) error {
	// Try to extract request ID and user ID from context
	var requestID *uuid.UUID
	var userID *uuid.UUID

	if reqID := ctx.Value("request_id"); reqID != nil {
		if id, ok := reqID.(uuid.UUID); ok {
			requestID = &id
		}
	}

	if uid := ctx.Value("user_id"); uid != nil {
		if id, ok := uid.(uuid.UUID); ok {
			userID = &id
		}
	}

	return s.errorLogger.LogError(ctx, LogErrorInput{
		Code:        code,
		Message:     message,
		Err:         err,
		Severity:    consts.SeverityError,
		ServiceName: s.serviceName,
		RequestID:   requestID,
		UserID:      userID,
		Metadata:    metadata,
	})
}

// LogDatabaseError logs a database error from service
func (s *ServiceErrorLogger) LogDatabaseError(ctx context.Context, err error, operation string, entity string) error {
	return s.LogError(ctx, consts.ErrCodeDatabaseError,
		fmt.Sprintf("Database %s failed for %s", operation, entity),
		err,
		map[string]any{
			"operation": operation,
			"entity":    entity,
		},
	)
}

// LogBusinessError logs a business logic error
func (s *ServiceErrorLogger) LogBusinessError(ctx context.Context, code string, message string, err error, details map[string]any) error {
	return s.LogError(ctx, code, message, err, details)
}
