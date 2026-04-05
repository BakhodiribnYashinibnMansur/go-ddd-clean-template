package errorx

import "testing"

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name string
		code string
		want ErrorSeverity
	}{
		// Critical
		{"ErrRepoDatabase is CRITICAL", ErrRepoDatabase, SeverityCritical},
		{"ErrRepoConnection is CRITICAL", ErrRepoConnection, SeverityCritical},
		{"ErrRepoTransaction is CRITICAL", ErrRepoTransaction, SeverityCritical},
		// High
		{"ErrUnauthorized is HIGH", ErrUnauthorized, SeverityHigh},
		{"ErrForbidden is HIGH", ErrForbidden, SeverityHigh},
		{"ErrPermissionDenied is HIGH", ErrPermissionDenied, SeverityHigh},
		// Medium
		{"ErrNotFound is MEDIUM", ErrNotFound, SeverityMedium},
		{"ErrUserNotFound is MEDIUM", ErrUserNotFound, SeverityMedium},
		{"ErrSessionNotFound is MEDIUM", ErrSessionNotFound, SeverityMedium},
		// Low
		{"ErrBadRequest is LOW", ErrBadRequest, SeverityLow},
		{"ErrInvalidInput is LOW", ErrInvalidInput, SeverityLow},
		{"ErrValidation is LOW", ErrValidation, SeverityLow},
		// Unknown defaults to MEDIUM
		{"unknown code defaults to MEDIUM", "UNKNOWN_CODE_XYZ", SeverityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSeverity(tt.code)
			if got != tt.want {
				t.Errorf("GetSeverity(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name string
		code string
		want ErrorCategory
	}{
		// VALIDATION
		{"ErrBadRequest is VALIDATION", ErrBadRequest, CategoryValidation},
		{"ErrInvalidInput is VALIDATION", ErrInvalidInput, CategoryValidation},
		{"ErrValidation is VALIDATION", ErrValidation, CategoryValidation},
		// SECURITY
		{"ErrUnauthorized is SECURITY", ErrUnauthorized, CategorySecurity},
		{"ErrForbidden is SECURITY", ErrForbidden, CategorySecurity},
		{"ErrPermissionDenied is SECURITY", ErrPermissionDenied, CategorySecurity},
		// DATA
		{"ErrNotFound is DATA", ErrNotFound, CategoryData},
		{"ErrUserNotFound is DATA", ErrUserNotFound, CategoryData},
		{"ErrSessionNotFound is DATA", ErrSessionNotFound, CategoryData},
		// BUSINESS
		{"ErrServiceBusinessRule is BUSINESS", ErrServiceBusinessRule, CategoryBusiness},
		// SYSTEM
		{"ErrRepoDatabase is SYSTEM", ErrRepoDatabase, CategorySystem},
		{"ErrTimeout is SYSTEM", ErrTimeout, CategorySystem},
		{"ErrRepoTimeout is SYSTEM", ErrRepoTimeout, CategorySystem},
		// EXTERNAL
		{"ErrServiceDependency is EXTERNAL", ErrServiceDependency, CategoryExternal},
		// Unknown defaults to SYSTEM
		{"unknown code defaults to SYSTEM", "UNKNOWN_CODE_XYZ", CategorySystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCategory(tt.code)
			if got != tt.want {
				t.Errorf("GetCategory(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetLayer(t *testing.T) {
	tests := []struct {
		name string
		code string
		want ErrorLayer
	}{
		{"HANDLER prefix returns HANDLER", "HANDLER_SOME_ERROR", LayerHandler},
		{"SERVICE prefix returns SERVICE", "SERVICE_SOME_ERROR", LayerService},
		{"REPO prefix returns REPOSITORY", "REPO_SOME_ERROR", LayerRepository},
		{"EXT_ prefix returns EXTERNAL", "EXT_SOME_ERROR", LayerExternal},
		{"unknown prefix returns SYSTEM", "OTHER_ERROR", LayerSystem},
		{"empty string returns SYSTEM", "", LayerSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLayer(tt.code)
			if got != tt.want {
				t.Errorf("GetLayer(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		// Retryable
		{"ErrTimeout is retryable", ErrTimeout, true},
		{"ErrRepoTimeout is retryable", ErrRepoTimeout, true},
		{"ErrRepoConnection is retryable", ErrRepoConnection, true},
		{"ErrServiceDependency is retryable", ErrServiceDependency, true},
		// Not retryable
		{"ErrBadRequest is not retryable", ErrBadRequest, false},
		{"ErrUnauthorized is not retryable", ErrUnauthorized, false},
		{"ErrNotFound is not retryable", ErrNotFound, false},
		{"ErrValidation is not retryable", ErrValidation, false},
		// Unknown
		{"unknown code is not retryable", "UNKNOWN_CODE_XYZ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.code)
			if got != tt.want {
				t.Errorf("IsRetryable(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetRetryStrategy(t *testing.T) {
	tests := []struct {
		name string
		code string
		want RetryStrategy
	}{
		{"ErrTimeout gets EXPONENTIAL", ErrTimeout, RetryStrategyExponential},
		{"ErrRepoTimeout gets EXPONENTIAL", ErrRepoTimeout, RetryStrategyExponential},
		{"ErrRepoConnection gets EXPONENTIAL", ErrRepoConnection, RetryStrategyExponential},
		{"ErrServiceDependency gets EXPONENTIAL", ErrServiceDependency, RetryStrategyExponential},
		{"ErrBadRequest gets NEVER", ErrBadRequest, RetryNever},
		{"ErrNotFound gets NEVER", ErrNotFound, RetryNever},
		{"unknown code gets NEVER", "UNKNOWN_CODE_XYZ", RetryNever},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRetryStrategy(tt.code)
			if got != tt.want {
				t.Errorf("GetRetryStrategy(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestShouldAlertOps(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		// Critical -> alert
		{"ErrRepoDatabase alerts ops", ErrRepoDatabase, true},
		{"ErrRepoConnection alerts ops", ErrRepoConnection, true},
		{"ErrRepoTransaction alerts ops", ErrRepoTransaction, true},
		// High -> alert
		{"ErrUnauthorized alerts ops", ErrUnauthorized, true},
		{"ErrForbidden alerts ops", ErrForbidden, true},
		{"ErrPermissionDenied alerts ops", ErrPermissionDenied, true},
		// Medium -> no alert
		{"ErrNotFound does not alert ops", ErrNotFound, false},
		{"ErrUserNotFound does not alert ops", ErrUserNotFound, false},
		// Low -> no alert
		{"ErrBadRequest does not alert ops", ErrBadRequest, false},
		{"ErrValidation does not alert ops", ErrValidation, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldAlertOps(tt.code)
			if got != tt.want {
				t.Errorf("ShouldAlertOps(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetErrorMetadata(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity ErrorSeverity
		category ErrorCategory
		layer    ErrorLayer
		retry    RetryStrategy
		retryable bool
		alertOps bool
	}{
		{
			name:     "ErrRepoDatabase metadata",
			code:     ErrRepoDatabase,
			severity: SeverityCritical,
			category: CategorySystem,
			layer:    LayerRepository,
			retry:    RetryNever,
			retryable: false,
			alertOps: true,
		},
		{
			name:     "ErrTimeout metadata",
			code:     ErrTimeout,
			severity: SeverityMedium,
			category: CategorySystem,
			layer:    LayerSystem,
			retry:    RetryStrategyExponential,
			retryable: true,
			alertOps: false,
		},
		{
			name:     "ErrBadRequest metadata",
			code:     ErrBadRequest,
			severity: SeverityLow,
			category: CategoryValidation,
			layer:    LayerSystem,
			retry:    RetryNever,
			retryable: false,
			alertOps: false,
		},
		{
			name:     "ErrServiceDependency metadata",
			code:     ErrServiceDependency,
			severity: SeverityMedium,
			category: CategoryExternal,
			layer:    LayerService,
			retry:    RetryStrategyExponential,
			retryable: true,
			alertOps: false,
		},
		{
			name:     "ErrUnauthorized metadata",
			code:     ErrUnauthorized,
			severity: SeverityHigh,
			category: CategorySecurity,
			layer:    LayerSystem,
			retry:    RetryNever,
			retryable: false,
			alertOps: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := GetErrorMetadata(tt.code)
			if m.Severity != tt.severity {
				t.Errorf("Severity = %q, want %q", m.Severity, tt.severity)
			}
			if m.Category != tt.category {
				t.Errorf("Category = %q, want %q", m.Category, tt.category)
			}
			if m.Layer != tt.layer {
				t.Errorf("Layer = %q, want %q", m.Layer, tt.layer)
			}
			if m.RetryStrategy != tt.retry {
				t.Errorf("RetryStrategy = %q, want %q", m.RetryStrategy, tt.retry)
			}
			if m.Retryable != tt.retryable {
				t.Errorf("Retryable = %v, want %v", m.Retryable, tt.retryable)
			}
			if m.AlertOps != tt.alertOps {
				t.Errorf("AlertOps = %v, want %v", m.AlertOps, tt.alertOps)
			}
			if !m.UserVisible {
				t.Error("UserVisible should be true by default")
			}
			if m.Tags == nil {
				t.Error("Tags should not be nil")
			}
			if m.CustomData == nil {
				t.Error("CustomData should not be nil")
			}
		})
	}
}
