package errors

import (
	"strings"
)

// External Telegram error codes
const (
	ErrExtTelegramAPIError    = "EXT_TELEGRAM_API_ERROR"
	CodeExtTelegramAPIError   = "7011"

	ErrExtTelegramTimeout     = "EXT_TELEGRAM_TIMEOUT"
	CodeExtTelegramTimeout    = "7012"

	ErrExtTelegramConnection  = "EXT_TELEGRAM_CONNECTION"
	CodeExtTelegramConnection = "7013"

	ErrExtTelegramRateLimit   = "EXT_TELEGRAM_RATE_LIMIT"
	CodeExtTelegramRateLimit  = "7014"
)

// HandleTelegramError converts Telegram API errors to AppError.
func HandleTelegramError(err error, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	var appErr *AppError

	switch {
	case strings.Contains(errMsg, "429") || strings.Contains(errMsg, "Too Many Requests"):
		appErr = AutoSource(New(ErrExtTelegramRateLimit, ""))
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		appErr = AutoSource(New(ErrExtTelegramTimeout, ""))
	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "dial"):
		appErr = AutoSource(New(ErrExtTelegramConnection, ""))
	default:
		appErr = AutoSource(Wrap(err, ErrExtTelegramAPIError, ""))
	}

	for k, v := range extraFields {
		_ = appErr.WithField(k, v)
	}
	return appErr
}
