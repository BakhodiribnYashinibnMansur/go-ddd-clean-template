package errors

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	mu sync.RWMutex

	// Error counts by code
	ErrorCounts map[string]int64

	// Error counts by severity
	SeverityCounts map[ErrorSeverity]int64

	// Error counts by category
	CategoryCounts map[ErrorCategory]int64

	// Total errors
	TotalErrors int64

	// Last error time
	LastErrorTime time.Time

	// Error rate (errors per minute)
	ErrorRate float64
}

// NewErrorMetrics creates a new ErrorMetrics instance
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ErrorCounts:    make(map[string]int64),
		SeverityCounts: make(map[ErrorSeverity]int64),
		CategoryCounts: make(map[ErrorCategory]int64),
	}
}

// RecordError records an error occurrence
func (m *ErrorMetrics) RecordError(err *AppError) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment total
	m.TotalErrors++

	// Increment by code
	m.ErrorCounts[err.Type]++

	// Increment by severity
	severity := err.GetSeverity()
	m.SeverityCounts[severity]++

	// Increment by category
	category := err.GetCategory()
	m.CategoryCounts[category]++

	// Update last error time
	m.LastErrorTime = time.Now()
}

// GetStats returns current error statistics
func (m *ErrorMetrics) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"total_errors":    m.TotalErrors,
		"error_counts":    m.ErrorCounts,
		"severity_counts": m.SeverityCounts,
		"category_counts": m.CategoryCounts,
		"last_error_time": m.LastErrorTime,
		"error_rate":      m.ErrorRate,
	}
}

// Reset resets all metrics
func (m *ErrorMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ErrorCounts = make(map[string]int64)
	m.SeverityCounts = make(map[ErrorSeverity]int64)
	m.CategoryCounts = make(map[ErrorCategory]int64)
	m.TotalErrors = 0
	m.ErrorRate = 0
}

// ErrorHook is a function that gets called when an error occurs
type ErrorHook func(ctx context.Context, err *AppError)

// ErrorHookManager manages error hooks
type ErrorHookManager struct {
	mu    sync.RWMutex
	hooks []ErrorHook
}

// NewErrorHookManager creates a new ErrorHookManager
func NewErrorHookManager() *ErrorHookManager {
	return &ErrorHookManager{
		hooks: make([]ErrorHook, 0),
	}
}

// AddHook adds an error hook
func (m *ErrorHookManager) AddHook(hook ErrorHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// ExecuteHooks executes all registered hooks
func (m *ErrorHookManager) ExecuteHooks(ctx context.Context, err *AppError) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.hooks {
		// Execute hook in goroutine to avoid blocking
		go func(h ErrorHook) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic in hook execution but don't propagate
				}
			}()
			h(ctx, err)
		}(hook)
	}
}

// Global error metrics and hooks
var (
	globalMetrics   *ErrorMetrics
	globalHookMgr   *ErrorHookManager
	metricsOnce     sync.Once
	hookManagerOnce sync.Once
)

// GetGlobalMetrics returns the global error metrics instance
func GetGlobalMetrics() *ErrorMetrics {
	metricsOnce.Do(func() {
		globalMetrics = NewErrorMetrics()
	})
	return globalMetrics
}

// GetGlobalHookManager returns the global error hook manager
func GetGlobalHookManager() *ErrorHookManager {
	hookManagerOnce.Do(func() {
		globalHookMgr = NewErrorHookManager()
	})
	return globalHookMgr
}

// RecordErrorGlobal records an error in global metrics and executes hooks
func RecordErrorGlobal(ctx context.Context, err *AppError) {
	if err == nil {
		return
	}

	// Record metrics
	GetGlobalMetrics().RecordError(err)

	// Execute hooks
	GetGlobalHookManager().ExecuteHooks(ctx, err)
}

// LoggingHook creates a hook that logs errors
func LoggingHook(logger *zap.Logger) ErrorHook {
	return func(ctx context.Context, err *AppError) {
		meta := err.GetMetadata()

		fields := []zap.Field{
			zap.String("error_type", err.Type),
			zap.String("error_code", err.Code),
			zap.String("severity", string(meta.Severity)),
			zap.String("category", string(meta.Category)),
			zap.Int("http_status", err.HTTPStatus),
		}

		if err.Details != "" {
			fields = append(fields, zap.String("details", err.Details))
		}

		// Add custom fields
		for key, value := range err.Fields {
			fields = append(fields, zap.Any(key, value))
		}

		// Log based on severity
		switch meta.Severity {
		case SeverityCritical:
			logger.Error(err.Message, fields...)
		case SeverityHigh:
			logger.Error(err.Message, fields...)
		case SeverityMedium:
			logger.Warn(err.Message, fields...)
		case SeverityLow:
			logger.Info(err.Message, fields...)
		default:
			logger.Info(err.Message, fields...)
		}
	}
}

// AlertingHook creates a hook that sends alerts for critical errors
func AlertingHook(alertFn func(ctx context.Context, err *AppError)) ErrorHook {
	return func(ctx context.Context, err *AppError) {
		meta := err.GetMetadata()

		// Only alert for critical and high severity errors
		if meta.Severity == SeverityCritical || meta.Severity == SeverityHigh {
			alertFn(ctx, err)
		}
	}
}

// MetricsHook creates a hook that records metrics
func MetricsHook(recordFn func(code string, severity ErrorSeverity, category ErrorCategory)) ErrorHook {
	return func(ctx context.Context, err *AppError) {
		meta := err.GetMetadata()
		recordFn(err.Type, meta.Severity, meta.Category)
	}
}
