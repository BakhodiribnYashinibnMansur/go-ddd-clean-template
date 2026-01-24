package metric

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"

	"go.uber.org/zap"
)

// MeasureSafe is a helper to measure function execution time and catch panics, saving results to DB.
// MeasureSafe is a helper to measure function execution time and catch panics, saving results to DB.
func (uc *UseCase) MeasureSafe(ctx context.Context, name string) func() {
	start := time.Now()

	// Pre-log start if needed, but usually we log/save at the end

	return func() {
		latency := time.Since(start)
		var panicErr *string
		isPanic := false

		if r := recover(); r != nil {
			isPanic = true
			errMsg := fmt.Sprintf("%v", r)
			panicErr = &errMsg

			// Log panic
			uc.logger.WithContext(ctx).Errorw("panic recovered in function", "func", name, "error", r, "latency_ms", latency.Milliseconds())

			// Re-panic after saving? Usually yes for "Safe" wrappers unless we want to suppress it.
			// The user's snippet does panic(r).
			defer panic(r)
		}

		// Save to DB
		// Create detached context because outer context might be cancelled
		saveCtx, cancel := context.WithTimeout(context.Background(), consts.DurationAuditSave*time.Second)
		defer cancel()

		metric := &domain.FunctionMetric{
			Name:       name,
			LatencyMs:  int(latency.Milliseconds()),
			IsPanic:    isPanic,
			PanicError: panicErr,
			CreatedAt:  time.Now(),
		}

		err := uc.Create(saveCtx, metric)
		if err != nil {
			uc.logger.WithContext(saveCtx).Errorw("failed to save function metric", "func", name, zap.Error(err))
		} else {
			uc.logger.WithContext(saveCtx).Infow("function execution tracked", "func", name, "latency_ms", latency.Milliseconds())
		}
	}
}
