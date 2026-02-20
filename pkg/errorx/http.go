package errorx

import (
	"context"
	"net/http"

	"gct/consts"
	"gct/pkg/logger"

	"github.com/google/uuid"
)

// HTTPErrorLogger provides HTTP-specific error logging functionality
type HTTPErrorLogger struct {
	errorLogger *ErrorLogger
	logger      logger.Log
}

// NewHTTPErrorLogger creates a new HTTP error logger
func NewHTTPErrorLogger(repo Repository, logger logger.Log) *HTTPErrorLogger {
	return &HTTPErrorLogger{
		errorLogger: NewErrorLogger(repo, logger),
		logger:      logger,
	}
}

// HTTPErrorContext contains HTTP request context for error logging
type HTTPErrorContext struct {
	RequestID *uuid.UUID
	UserID    *uuid.UUID
	IPAddress string
	Path      string
	Method    string
}

// LogHTTPError logs an HTTP error with full context
func (h *HTTPErrorLogger) LogHTTPError(ctx context.Context, code string, message string, err error, httpCtx HTTPErrorContext, metadata map[string]any) error {
	return h.errorLogger.LogError(ctx, LogErrorInput{
		Code:        code,
		Message:     message,
		Err:         err,
		Severity:    consts.SeverityError,
		ServiceName: "api",
		RequestID:   httpCtx.RequestID,
		UserID:      httpCtx.UserID,
		IPAddress:   &httpCtx.IPAddress,
		Path:        &httpCtx.Path,
		Method:      &httpCtx.Method,
		Metadata:    metadata,
	})
}

// LogAuthError logs an authentication error
func (h *HTTPErrorLogger) LogAuthError(ctx context.Context, err error, httpCtx HTTPErrorContext, username string) error {
	metadata := map[string]any{
		"username": username,
	}

	return h.LogHTTPError(ctx, consts.ErrCodeAuthFailed, "Authentication failed", err, httpCtx, metadata)
}

// LogValidationError logs a validation error
func (h *HTTPErrorLogger) LogValidationError(ctx context.Context, err error, httpCtx HTTPErrorContext, field string, value any) error {
	metadata := map[string]any{
		"field": field,
		"value": value,
	}

	return h.LogHTTPError(ctx, consts.ErrCodeValidationFailed, "Validation failed", err, httpCtx, metadata)
}

// LogDatabaseError logs a database error
func (h *HTTPErrorLogger) LogDatabaseError(ctx context.Context, err error, httpCtx HTTPErrorContext, operation string, table string) error {
	metadata := map[string]any{
		"operation": operation,
		"table":     table,
	}

	return h.LogHTTPError(ctx, consts.ErrCodeDatabaseError, "Database operation failed", err, httpCtx, metadata)
}

// LogExternalServiceError logs an external service error
func (h *HTTPErrorLogger) LogExternalServiceError(ctx context.Context, err error, httpCtx HTTPErrorContext, service string, endpoint string) error {
	metadata := map[string]any{
		"service":  service,
		"endpoint": endpoint,
	}

	return h.LogHTTPError(ctx, consts.ErrCodeExternalServiceError, "External service call failed", err, httpCtx, metadata)
}

// ExtractHTTPContext extracts HTTP context from request
func ExtractHTTPContext(r *http.Request) HTTPErrorContext {
	ctx := HTTPErrorContext{
		Path:   r.URL.Path,
		Method: r.Method,
	}

	// Extract request ID from context (if available)
	if reqID := r.Context().Value("request_id"); reqID != nil {
		if id, ok := reqID.(uuid.UUID); ok {
			ctx.RequestID = &id
		} else if idStr, ok := reqID.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				ctx.RequestID = &id
			}
		}
	}

	// Extract user ID from context (if available)
	if userID := r.Context().Value("user_id"); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			ctx.UserID = &id
		} else if idStr, ok := userID.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				ctx.UserID = &id
			}
		}
	}

	// Extract IP address
	ctx.IPAddress = getClientIP(r)

	return ctx
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Example usage in HTTP handler:
/*
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	httpCtx := errorx.ExtractHTTPContext(r)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorLogger.LogValidationError(ctx, err, httpCtx, "body", "invalid JSON")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.useCase.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		h.errorLogger.LogAuthError(ctx, err, httpCtx, req.Username)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Success...
}
*/
