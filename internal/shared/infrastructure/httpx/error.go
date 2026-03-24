// Package util contains helper functions and categorized error instances
// used across the restapi layer to simplify request processing and error handling.
package httpx

import (
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	"go.uber.org/zap"
)

// LogError provides a standardized way to log handler-layer errors using zap fields,
// automatically tagging the log entry with a consistent error type.
func LogError(l logger.Log, err error, msg string) {
	l.Errorw(msg,
		zap.Error(err),
		zap.String("type", consts.TypeHandlerError),
	)
}

// Global categorized error instances for reuse in Gin handlers.
// Each error is mapped to a specific HTTP-equivalent status through apperrors.
var (
	// Validation / Bad Request (400) - Errors related to malformed or invalid client input.
	ErrParamIsEmpty     = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "parameter is empty")
	ErrParsingQuery     = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "error while parsing query parameter")
	ErrUnmarshalData    = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "error while unmarshalling data")
	ErrParamIsInvalid   = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "parameter is invalid")
	ErrParsingUUID      = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "error while parsing UUID")
	ErrDataUnsignedInt  = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "value is not a positive integer")
	ErrDateOrderEmpty   = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "date-order query is empty")
	ErrDateOrderInvalid = apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "invalid date-order value. it is not same with asc or desc")

	// Authentication (401) - Errors related to identity verification and token lifecycle.
	ErrUserIdNotFound        = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "user id not found in context")
	ErrSessionIDNotFound     = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "sessionID not found")
	ErrInvalidSessionID      = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "sessionID is not a string or UUID")
	ErrApiKeyTypeNotFound    = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "API key type not found")
	ErrSessionNotFound       = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "session object not found in context")
	ErrSessionCastFailed     = apperrors.NewHandlerError(apperrors.ErrHandlerInternal, "failed to cast session object")
	ErrUnAuth                = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "unauthorized. token is missing")
	ErrInvalidToken          = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "unauthorized. token is invalid")
	ErrExpiredToken          = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "unauthorized. token is expired")
	ErrRevokedToken          = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "unauthorized. token is revoked")
	ErrInvalidIssuer         = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid issuer")
	ErrInvalidType           = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid token type")
	ErrInvalidSession        = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid session id in token")
	ErrInvalidRefreshFormat  = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid refresh token format")
	ErrInvalidRefreshToken   = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid refresh token")
	ErrInvalidRefreshSession = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid refresh session")
	ErrApiKeyMissing         = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "API key missing")
	ErrInvalidApiKey         = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "invalid API key")
	ErrUserNotFound          = apperrors.NewHandlerError(apperrors.ErrHandlerUnauthorized, "user not found")

	// Forbidden (403) - Errors related to authorization, CSRF protection, and security policies.
	ErrRoleNotFound            = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "user role not found in context")
	ErrCSRFMissing             = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "CSRF token is missing")
	ErrCSRFInvalid             = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "CSRF token is invalid or mismatched")
	ErrAccessDenied            = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "access denied")
	ErrFetchMetadataSuspicious = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "Request blocked: suspicious source (Postman/cURL blocked in production)")
	ErrFetchMetadataBlocked    = apperrors.NewHandlerError(apperrors.ErrHandlerForbidden, "Request blocked by Fetch Metadata policy (cross-site requests blocked for security)")

	// Rate Limit (429) - Triggered when client exceeds allowed request thresholds.
	ErrRateLimitExceeded = apperrors.NewHandlerError(apperrors.ErrHandlerTooManyRequests, "too many requests")

	// Internal (500) - Unhandled exceptions, infrastructure failures, or application panics.
	ErrInternalError     = apperrors.NewHandlerError(apperrors.ErrHandlerInternal, "internal server error")
	ErrRateLimitInternal = apperrors.NewHandlerError(apperrors.ErrHandlerInternal, "rate limiter internal error")
	ErrPanicRecovered    = apperrors.NewHandlerError(apperrors.ErrHandlerInternal, "internal server error (panic)")
)

const (
	// Templates for formatting dynamic error messages.
	ParamInvalid = "parameter %s is invalid"
	QueryInvalid = "query parameter %s is invalid"
)
