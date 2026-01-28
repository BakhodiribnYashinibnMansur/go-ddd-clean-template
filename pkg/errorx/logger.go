package errorx

import (
	"context"

	"gct/pkg/logger"

	"github.com/google/uuid"
)

// Repository defines the interface for storing system errors
type Repository interface {
	Create(ctx context.Context, input LogErrorInput) error
}

// ErrorLogger provides functionality to log errors to a repository
type ErrorLogger struct {
	repo   Repository
	logger logger.Log
}

// NewErrorLogger creates a new error logger instance
func NewErrorLogger(repo Repository, logger logger.Log) *ErrorLogger {
	return &ErrorLogger{
		repo:   repo,
		logger: logger,
	}
}

// LogErrorInput represents input for logging an error
type LogErrorInput struct {
	Code        string
	Message     string
	Err         error
	Severity    string // ERROR, FATAL, PANIC, WARN
	ServiceName string
	RequestID   *uuid.UUID
	UserID      *uuid.UUID
	IPAddress   *string
	Path        *string
	Method      *string
	Metadata    map[string]interface{}
}

// LogError logs an error to the repository
func (e *ErrorLogger) LogError(ctx context.Context, input LogErrorInput) error {
	// Set default severity
	if input.Severity == "" {
		input.Severity = "ERROR"
	}

	// Log to repository
	err := e.repo.Create(ctx, input)
	if err != nil {
		e.logger.Error("failed to log error to repository",
			"error", err,
			"code", input.Code,
			"message", input.Message,
		)
		return err
	}

	return nil
}

// LogErrorSimple logs a simple error with minimal information
func (e *ErrorLogger) LogErrorSimple(ctx context.Context, code string, message string, err error) error {
	return e.LogError(ctx, LogErrorInput{
		Code:     code,
		Message:  message,
		Err:      err,
		Severity: "ERROR",
	})
}

// LogFatal logs a fatal error
func (e *ErrorLogger) LogFatal(ctx context.Context, code string, message string, err error) error {
	return e.LogError(ctx, LogErrorInput{
		Code:     code,
		Message:  message,
		Err:      err,
		Severity: "FATAL",
	})
}

// LogPanic logs a panic error
func (e *ErrorLogger) LogPanic(ctx context.Context, code string, message string, err error) error {
	return e.LogError(ctx, LogErrorInput{
		Code:     code,
		Message:  message,
		Err:      err,
		Severity: "PANIC",
	})
}

// LogWarn logs a warning
func (e *ErrorLogger) LogWarn(ctx context.Context, code string, message string, err error) error {
	return e.LogError(ctx, LogErrorInput{
		Code:     code,
		Message:  message,
		Err:      err,
		Severity: "WARN",
	})
}
