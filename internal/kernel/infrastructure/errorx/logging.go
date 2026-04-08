package errorx

import (
	"errors"
	"sync/atomic"

	"go.uber.org/zap"
)

type Reporter interface {
	SendError(err error) error
}

var reporterPtr atomic.Pointer[Reporter]

func SetReporter(r Reporter) {
	reporterPtr.Store(&r)
}

func getReporter() Reporter {
	if p := reporterPtr.Load(); p != nil {
		return *p
	}
	return nil
}

// LogError logs error using zap logger with all available fields
// Usage in Handler layer ONLY:
//
//	errors.LogError(logger, err)
func LogError(logger *zap.Logger, err error) {
	if r := getReporter(); r != nil {
		_ = r.SendError(err)
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		// Not our custom error, log as standard error
		logger.Error("error occurred", zap.Error(err))
		return
	}

	fields := []zap.Field{
		zap.String("error_type", appErr.Type),
		zap.String("error_code", appErr.Code),
		zap.Int("http_status", appErr.HTTPStatus),
		zap.String("user_message", appErr.UserMsg),
	}

	// Add details if present
	if appErr.Details != "" {
		fields = append(fields, zap.String("details", appErr.Details))
	}

	// Add custom fields
	for key, value := range appErr.Fields {
		fields = append(fields, zap.Any(key, value))
	}

	// Add wrapped error if present
	if appErr.Err != nil {
		fields = append(fields, zap.NamedError("wrapped_error", appErr.Err))
	}

	logger.Error(appErr.Message, fields...)
}

// LogWarn logs error as warning
func LogWarn(logger *zap.Logger, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		logger.Warn("warning occurred", zap.Error(err))
		return
	}

	fields := []zap.Field{
		zap.String("error_type", appErr.Type),
		zap.String("error_code", appErr.Code),
		zap.String("message", appErr.Message),
	}

	for key, value := range appErr.Fields {
		fields = append(fields, zap.Any(key, value))
	}

	logger.Warn(appErr.UserMsg, fields...)
}

// LogInfo logs error information without full stack trace
func LogInfo(logger *zap.Logger, err error, message string) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		logger.Info(message, zap.Error(err))
		return
	}

	logger.Info(message,
		zap.String("error_type", appErr.Type),
		zap.String("error_code", appErr.Code),
		zap.String("user_message", appErr.UserMsg),
	)
}
