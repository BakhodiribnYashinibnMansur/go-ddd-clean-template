package errors

import "strings"

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "CRITICAL" // System-wide impact, immediate action required
	SeverityHigh     ErrorSeverity = "HIGH"     // Major functionality affected
	SeverityMedium   ErrorSeverity = "MEDIUM"   // Partial functionality affected
	SeverityLow      ErrorSeverity = "LOW"      // Minor issue, no significant impact
	SeverityInfo     ErrorSeverity = "INFO"     // Informational, not an actual error
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	CategoryValidation ErrorCategory = "VALIDATION" // Input validation errors
	CategorySecurity   ErrorCategory = "SECURITY"   // Authentication/Authorization errors
	CategoryData       ErrorCategory = "DATA"       // Data access/persistence errors
	CategoryBusiness   ErrorCategory = "BUSINESS"   // Business logic violations
	CategorySystem     ErrorCategory = "SYSTEM"     // System/infrastructure errors
	CategoryExternal   ErrorCategory = "EXTERNAL"   // External service errors
	CategoryNetwork    ErrorCategory = "NETWORK"    // Network/connectivity errors
)

// ErrorLayer represents which layer the error originated from
type ErrorLayer string

const (
	LayerHandler    ErrorLayer = "HANDLER"    // HTTP/Controller layer
	LayerService    ErrorLayer = "SERVICE"    // Business logic layer
	LayerRepository ErrorLayer = "REPOSITORY" // Data access layer
	LayerExternal   ErrorLayer = "EXTERNAL"   // External service integration
	LayerSystem     ErrorLayer = "SYSTEM"     // System/infrastructure
)

// RetryStrategy defines how an error should be handled for retries
type RetryStrategy string

const (
	RetryNever               RetryStrategy = "NEVER"       // Never retry this error
	RetryStrategyImmediate   RetryStrategy = "IMMEDIATE"   // Retry immediately
	RetryStrategyExponential RetryStrategy = "EXPONENTIAL" // Retry with exponential backoff
	RetryStrategyLinear      RetryStrategy = "LINEAR"      // Retry with linear backoff
)

// ErrorMetadata contains additional metadata about an error
type ErrorMetadata struct {
	Severity      ErrorSeverity  `json:"severity"`
	Category      ErrorCategory  `json:"category"`
	Layer         ErrorLayer     `json:"layer"`
	RetryStrategy RetryStrategy  `json:"retry_strategy"`
	Retryable     bool           `json:"retryable"`
	UserVisible   bool           `json:"user_visible"` // Whether to show to end users
	AlertOps      bool           `json:"alert_ops"`    // Whether to alert operations team
	Tags          []string       `json:"tags,omitempty"`
	CustomData    map[string]any `json:"custom_data,omitempty"`
}

// GetSeverity returns the severity level for an error code
func GetSeverity(code string) ErrorSeverity {
	// Check domain/external codes first
	if s, ok := domainSeverities[code]; ok {
		return s
	}

	switch code {
	// Critical errors
	case ErrRepoDatabase, ErrRepoConnection, ErrRepoTransaction:
		return SeverityCritical

	// High severity
	case ErrUnauthorized, ErrForbidden, ErrPermissionDenied, ErrServicePolicyViolation:
		return SeverityHigh

	// Medium severity
	case ErrNotFound, ErrUserNotFound, ErrSessionNotFound, ErrServiceNotFound:
		return SeverityMedium

	// Low severity
	case ErrBadRequest, ErrInvalidInput, ErrValidation:
		return SeverityLow

	default:
		return SeverityMedium
	}
}

// GetCategory returns the category for an error code
func GetCategory(code string) ErrorCategory {
	// Check domain/external codes first
	if c, ok := domainCategories[code]; ok {
		return c
	}

	switch code {
	case ErrBadRequest, ErrInvalidInput, ErrValidation, ErrServiceInvalidInput, ErrServiceValidation:
		return CategoryValidation

	case ErrUnauthorized, ErrInvalidToken, ErrExpiredToken, ErrRevokedToken,
		ErrForbidden, ErrPermissionDenied, ErrDisabledAccount, ErrServicePolicyViolation:
		return CategorySecurity

	case ErrNotFound, ErrUserNotFound, ErrSessionNotFound, ErrConflict, ErrAlreadyExists,
		ErrRepoNotFound, ErrRepoAlreadyExists, ErrRepoConstraint:
		return CategoryData

	case ErrServiceBusinessRule, ErrServiceConflict:
		return CategoryBusiness

	case ErrInternal, ErrDatabase, ErrTimeout, ErrRepoDatabase, ErrRepoTimeout,
		ErrRepoConnection, ErrRepoTransaction:
		return CategorySystem

	case ErrServiceDependency:
		return CategoryExternal

	default:
		return CategorySystem
	}
}

// GetLayer returns the layer for an error code
func GetLayer(code string) ErrorLayer {
	switch {
	case strings.HasPrefix(code, "HANDLER"):
		return LayerHandler
	case strings.HasPrefix(code, "SERVICE"):
		return LayerService
	case strings.HasPrefix(code, "REPO"):
		return LayerRepository
	case strings.HasPrefix(code, "EXT_"):
		return LayerExternal
	default:
		return LayerSystem
	}
}

// IsRetryable determines if an error is retryable
func IsRetryable(code string) bool {
	switch code {
	// Retryable errors
	case ErrTimeout, ErrRepoTimeout, ErrRepoConnection, ErrServiceDependency:
		return true

	// Non-retryable errors
	case ErrBadRequest, ErrInvalidInput, ErrValidation, ErrUnauthorized,
		ErrForbidden, ErrNotFound, ErrConflict, ErrAlreadyExists:
		return false

	default:
		return false
	}
}

// GetRetryStrategy returns the retry strategy for an error code
func GetRetryStrategy(code string) RetryStrategy {
	if !IsRetryable(code) {
		return RetryNever
	}

	switch code {
	case ErrTimeout, ErrRepoTimeout:
		return RetryStrategyExponential
	case ErrRepoConnection, ErrServiceDependency:
		return RetryStrategyExponential
	default:
		return RetryNever
	}
}

// ShouldAlertOps determines if operations team should be alerted
func ShouldAlertOps(code string) bool {
	severity := GetSeverity(code)
	return severity == SeverityCritical || severity == SeverityHigh
}

// GetErrorMetadata returns complete metadata for an error code
func GetErrorMetadata(code string) ErrorMetadata {
	return ErrorMetadata{
		Severity:      GetSeverity(code),
		Category:      GetCategory(code),
		Layer:         GetLayer(code),
		RetryStrategy: GetRetryStrategy(code),
		Retryable:     IsRetryable(code),
		UserVisible:   true, // Most errors are user-visible by default
		AlertOps:      ShouldAlertOps(code),
		Tags:          []string{},
		CustomData:    make(map[string]any),
	}
}
