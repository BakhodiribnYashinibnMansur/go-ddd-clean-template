package errors

import (
	"strings"
)

// External Asynq error codes
const (
	ErrExtAsynqEnqueueFailed  = "EXT_ASYNQ_ENQUEUE_FAILED"
	CodeExtAsynqEnqueueFailed = "7021"

	ErrExtAsynqConnection     = "EXT_ASYNQ_CONNECTION"
	CodeExtAsynqConnection    = "7022"

	ErrExtAsynqTimeout        = "EXT_ASYNQ_TIMEOUT"
	CodeExtAsynqTimeout       = "7023"

	ErrExtAsynqPayloadError   = "EXT_ASYNQ_PAYLOAD_ERROR"
	CodeExtAsynqPayloadError  = "7024"
)

// HandleAsynqError converts Asynq task queue errors to AppError.
func HandleAsynqError(err error, taskType string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	var appErr *AppError

	switch {
	case strings.Contains(errMsg, "marshal"):
		appErr = AutoSource(Wrap(err, ErrExtAsynqPayloadError, ""))
	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "dial") || strings.Contains(errMsg, "EOF"):
		appErr = AutoSource(New(ErrExtAsynqConnection, ""))
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		appErr = AutoSource(New(ErrExtAsynqTimeout, ""))
	default:
		appErr = AutoSource(Wrap(err, ErrExtAsynqEnqueueFailed, ""))
	}

	if taskType != "" {
		_ = appErr.WithField("task_type", taskType)
	}
	for k, v := range extraFields {
		_ = appErr.WithField(k, v)
	}
	return appErr
}
