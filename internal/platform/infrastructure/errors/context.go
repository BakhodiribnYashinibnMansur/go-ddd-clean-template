package errors

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	unknownValue = "unknown"
)

// ErrorContext provides additional context for errors
type ErrorContext struct {
	UserID    string `json:"user_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Operation  string         `json:"operation,omitempty"`
	Resource   string         `json:"resource,omitempty"`
	ResourceID string         `json:"resource_id,omitempty"`
	IPAddress  string         `json:"ip_address,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Path       string         `json:"path,omitempty"`
	Method     string         `json:"method,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ContextKey type for context keys
type ContextKey string

const (
	ContextKeyErrorContext ContextKey = "error_context"
	ContextKeyUserID       ContextKey = "user_id"
	ContextKeyRequestID ContextKey = "request_id"
)

// GetErrorContext extracts ErrorContext from context
func GetErrorContext(ctx context.Context) *ErrorContext {
	if ctx == nil {
		return &ErrorContext{
			Metadata: make(map[string]any),
		}
	}

	// Try to get existing error context
	if ec, ok := ctx.Value(ContextKeyErrorContext).(*ErrorContext); ok {
		return ec
	}

	// Build from individual context values
	ec := &ErrorContext{
		Metadata: make(map[string]any),
	}

	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		ec.UserID = userID
	}
	if reqID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		ec.RequestID = reqID
	}

	return ec
}

// WithErrorContext adds ErrorContext to context
func WithErrorContext(ctx context.Context, ec *ErrorContext) context.Context {
	return context.WithValue(ctx, ContextKeyErrorContext, ec)
}

// WithSource adds source file and function information to error
// This helps track where the error originated in the codebase
func WithSource(err *AppError, file, function string) *AppError {
	if err == nil {
		return nil
	}
	return err.
		WithField("file", file).
		WithField("function", function)
}

// GetCaller returns the file and function name of the caller
// Skip levels: 0 = GetCaller itself, 1 = caller of GetCaller, etc.
// Returns relative path from project root (internal/repo/...) and short function name
func GetCaller(skip int) (file, function string) {
	pc, fullPath, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return unknownValue, unknownValue
	}

	// Extract relative path from project root
	// Look for common project markers: internal/, pkg/, cmd/
	file = fullPath
	if idx := strings.Index(fullPath, "/internal/"); idx != -1 {
		file = fullPath[idx+1:] // Remove leading slash, keep "internal/..."
	} else if idx := strings.Index(fullPath, "/pkg/"); idx != -1 {
		file = fullPath[idx+1:]
	} else if idx := strings.Index(fullPath, "/cmd/"); idx != -1 {
		file = fullPath[idx+1:]
	} else {
		// If no marker found, just use the filename
		file = filepath.Base(fullPath)
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, unknownValue
	}

	// Extract short function name (e.g., "Repo.Get" instead of full path)
	funcName := fn.Name()
	// Split by "/" and take last part, then split by "." and take last 2 parts
	parts := strings.Split(funcName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Now get package.Function or Struct.Method
		dotParts := strings.Split(lastPart, ".")
		if len(dotParts) >= 2 {
			// Return last 2 parts: "Repo.Get" or "package.Function"
			funcName = strings.Join(dotParts[len(dotParts)-2:], ".")
		}
	}

	return file, funcName
}

// WithCaller automatically adds caller information to error
// Usage: errors.WithCaller(err, 0) - will get the caller of WithCaller
func WithCaller(err *AppError, skip int) *AppError {
	if err == nil {
		return nil
	}
	file, function := GetCaller(skip + 1)
	return err.
		WithField("file", file).
		WithField("function", function)
}

// AutoSource automatically adds caller info with proper skip level
// This should be called directly from repository/service functions
func AutoSource(err *AppError) *AppError {
	if err == nil {
		return nil
	}
	// skip = 1 means we skip AutoSource itself and get the actual caller
	file, function := GetCaller(1)
	return err.
		WithField("file", file).
		WithField("function", function)
}

// WithOperation adds operation name to error context
func (e *AppError) WithOperation(operation string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields["operation"] = operation
	return e
}

// WithResource adds resource information to error
func (e *AppError) WithResource(resource, resourceID string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields["resource"] = resource
	if resourceID != "" {
		e.Fields["resource_id"] = resourceID
	}
	return e
}

// WithContext enriches error with context information
func (e *AppError) WithContext(ctx context.Context) *AppError {
	ec := GetErrorContext(ctx)

	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}

	if ec.UserID != "" {
		e.Fields["user_id"] = ec.UserID
	}
	if ec.RequestID != "" {
		e.Fields["request_id"] = ec.RequestID
	}
	if ec.Operation != "" {
		e.Fields["operation"] = ec.Operation
	}
	if ec.Resource != "" {
		e.Fields["resource"] = ec.Resource
	}
	if ec.ResourceID != "" {
		e.Fields["resource_id"] = ec.ResourceID
	}
	if ec.IPAddress != "" {
		e.Fields["ip_address"] = ec.IPAddress
	}
	if ec.UserAgent != "" {
		e.Fields["user_agent"] = ec.UserAgent
	}
	if ec.Path != "" {
		e.Fields["path"] = ec.Path
	}
	if ec.Method != "" {
		e.Fields["method"] = ec.Method
	}

	// Add metadata
	for k, v := range ec.Metadata {
		e.Fields[k] = v
	}

	return e
}

// WithMetadata adds metadata to error
func (e *AppError) WithMetadata(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithTag adds a tag to error
func (e *AppError) WithTag(tag string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}

	tags, ok := e.Fields["tags"].([]string)
	if !ok {
		tags = []string{}
	}

	tags = append(tags, tag)
	e.Fields["tags"] = tags

	return e
}

// GetMetadata returns error metadata
func (e *AppError) GetMetadata() ErrorMetadata {
	meta := GetErrorMetadata(e.Type)

	// Use native fields if set
	if e.Severity != "" {
		meta.Severity = e.Severity
	}
	if e.Category != "" {
		meta.Category = e.Category
	}

	// Override with custom data from Fields if present
	if e.Fields != nil {
		if tags, ok := e.Fields["tags"].([]string); ok {
			meta.Tags = tags
		}
		meta.CustomData = e.Fields
	}

	return meta
}

// IsRetryable checks if this error is retryable
func (e *AppError) IsRetryable() bool {
	return IsRetryable(e.Type)
}

// GetSeverity returns the severity of this error
func (e *AppError) GetSeverity() ErrorSeverity {
	return GetSeverity(e.Type)
}

// GetCategory returns the category of this error
func (e *AppError) GetCategory() ErrorCategory {
	return GetCategory(e.Type)
}

// String returns a formatted string representation
func (e *AppError) String() string {
	meta := e.GetMetadata()
	return fmt.Sprintf("[%s][%s][%s] %s: %s",
		meta.Severity,
		meta.Category,
		e.Code,
		e.Type,
		e.Message,
	)
}
