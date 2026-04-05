package errorx

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
	errorCounts map[string]int64

	// Error counts by severity
	severityCounts map[ErrorSeverity]int64

	// Error counts by category
	categoryCounts map[ErrorCategory]int64

	// Total errors
	totalErrors int64

	// Last error time
	lastErrorTime time.Time

	// Error rate (errors per minute)
	errorRate float64
}

// NewErrorMetrics creates a new ErrorMetrics instance
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		errorCounts:    make(map[string]int64),
		severityCounts: make(map[ErrorSeverity]int64),
		categoryCounts: make(map[ErrorCategory]int64),
	}
}

// RecordError records an error occurrence
func (m *ErrorMetrics) RecordError(err *AppError) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalErrors++
	m.errorCounts[err.Type]++

	severity := err.GetSeverity()
	m.severityCounts[severity]++

	category := err.GetCategory()
	m.categoryCounts[category]++

	m.lastErrorTime = time.Now()
}

// GetStats returns current error statistics as deep copies (safe for concurrent use).
func (m *ErrorMetrics) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ec := make(map[string]int64, len(m.errorCounts))
	for k, v := range m.errorCounts {
		ec[k] = v
	}
	sc := make(map[ErrorSeverity]int64, len(m.severityCounts))
	for k, v := range m.severityCounts {
		sc[k] = v
	}
	cc := make(map[ErrorCategory]int64, len(m.categoryCounts))
	for k, v := range m.categoryCounts {
		cc[k] = v
	}

	return map[string]any{
		"total_errors":    m.totalErrors,
		"error_counts":    ec,
		"severity_counts": sc,
		"category_counts": cc,
		"last_error_time": m.lastErrorTime,
		"error_rate":      m.errorRate,
	}
}

// Reset resets all metrics
func (m *ErrorMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorCounts = make(map[string]int64)
	m.severityCounts = make(map[ErrorSeverity]int64)
	m.categoryCounts = make(map[ErrorCategory]int64)
	m.totalErrors = 0
	m.errorRate = 0
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
