package errors

import (
	"strings"
)

// External Firebase error codes
const (
	ErrExtFirebaseSendFailed     = "EXT_FIREBASE_SEND_FAILED"
	CodeExtFirebaseSendFailed    = "7001"

	ErrExtFirebaseInvalidToken   = "EXT_FIREBASE_INVALID_TOKEN"
	CodeExtFirebaseInvalidToken  = "7002"

	ErrExtFirebaseQuotaExceeded  = "EXT_FIREBASE_QUOTA_EXCEEDED"
	CodeExtFirebaseQuotaExceeded = "7003"

	ErrExtFirebaseUnavailable    = "EXT_FIREBASE_UNAVAILABLE"
	CodeExtFirebaseUnavailable   = "7004"
)

// HandleFirebaseError converts Firebase/FCM errors to AppError.
func HandleFirebaseError(err error, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	var appErr *AppError

	switch {
	case strings.Contains(errMsg, "registration-token-not-registered") ||
		strings.Contains(errMsg, "invalid-registration-token"):
		appErr = AutoSource(New(ErrExtFirebaseInvalidToken, ""))

	case strings.Contains(errMsg, "quota-exceeded") ||
		strings.Contains(errMsg, "message-rate-exceeded"):
		appErr = AutoSource(New(ErrExtFirebaseQuotaExceeded, ""))

	case strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "internal-error") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline"):
		appErr = AutoSource(New(ErrExtFirebaseUnavailable, ""))

	default:
		appErr = AutoSource(Wrap(err, ErrExtFirebaseSendFailed, ""))
	}

	for k, v := range extraFields {
		_ = appErr.WithField(k, v)
	}
	return appErr
}
