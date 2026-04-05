package errors

import (
	"strings"
)

// External EventBus error codes
const (
	ErrExtEventBusPublishFailed   = "EXT_EVENTBUS_PUBLISH_FAILED"
	CodeExtEventBusPublishFailed  = "7031"

	ErrExtEventBusConnection      = "EXT_EVENTBUS_CONNECTION"
	CodeExtEventBusConnection     = "7032"

	ErrExtEventBusTimeout         = "EXT_EVENTBUS_TIMEOUT"
	CodeExtEventBusTimeout        = "7033"

	ErrExtEventBusSubscribeFailed  = "EXT_EVENTBUS_SUBSCRIBE_FAILED"
	CodeExtEventBusSubscribeFailed = "7034"
)

// HandleEventBusError converts EventBus (Redis Streams) errors to AppError.
func HandleEventBusError(err error, channel string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	var appErr *AppError

	switch {
	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "EOF") || strings.Contains(errMsg, "broken pipe"):
		appErr = AutoSource(New(ErrExtEventBusConnection, ""))
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		appErr = AutoSource(New(ErrExtEventBusTimeout, ""))
	default:
		appErr = AutoSource(Wrap(err, ErrExtEventBusPublishFailed, ""))
	}

	if channel != "" {
		_ = appErr.WithField("channel", channel)
	}
	for k, v := range extraFields {
		_ = appErr.WithField(k, v)
	}
	return appErr
}
