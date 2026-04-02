package errors

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewErrorMetrics(t *testing.T) {
	m := NewErrorMetrics()
	if m == nil {
		t.Fatal("expected non-nil ErrorMetrics")
	}
	if m.errorCounts == nil {
		t.Error("expected initialized ErrorCounts map")
	}
	if m.severityCounts == nil {
		t.Error("expected initialized SeverityCounts map")
	}
	if m.categoryCounts == nil {
		t.Error("expected initialized CategoryCounts map")
	}
	if m.totalErrors != 0 {
		t.Errorf("expected TotalErrors 0, got %d", m.totalErrors)
	}
}

func TestErrorMetrics_RecordError(t *testing.T) {
	m := NewErrorMetrics()

	err1 := New(ErrBadRequest, "bad input")
	err2 := New(ErrNotFound, "missing")
	err3 := New(ErrBadRequest, "another bad input")

	m.RecordError(err1)
	m.RecordError(err2)
	m.RecordError(err3)

	if m.totalErrors != 3 {
		t.Errorf("expected TotalErrors 3, got %d", m.totalErrors)
	}

	if m.errorCounts[ErrBadRequest] != 2 {
		t.Errorf("expected BAD_REQUEST count 2, got %d", m.errorCounts[ErrBadRequest])
	}
	if m.errorCounts[ErrNotFound] != 1 {
		t.Errorf("expected NOT_FOUND count 1, got %d", m.errorCounts[ErrNotFound])
	}

	if m.lastErrorTime.IsZero() {
		t.Error("expected LastErrorTime to be set")
	}
}

func TestErrorMetrics_RecordError_Concurrent(t *testing.T) {
	m := NewErrorMetrics()
	var wg sync.WaitGroup
	count := 100

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			m.RecordError(New(ErrInternal, "concurrent error"))
		}()
	}
	wg.Wait()

	if m.totalErrors != int64(count) {
		t.Errorf("expected TotalErrors %d, got %d", count, m.totalErrors)
	}
}

func TestErrorMetrics_GetStats(t *testing.T) {
	m := NewErrorMetrics()
	m.RecordError(New(ErrBadRequest, "test"))

	stats := m.GetStats()
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}

	total, ok := stats["total_errors"].(int64)
	if !ok {
		t.Fatal("expected total_errors to be int64")
	}
	if total != 1 {
		t.Errorf("expected total_errors 1, got %d", total)
	}

	if _, ok := stats["error_counts"]; !ok {
		t.Error("expected error_counts in stats")
	}
	if _, ok := stats["severity_counts"]; !ok {
		t.Error("expected severity_counts in stats")
	}
	if _, ok := stats["category_counts"]; !ok {
		t.Error("expected category_counts in stats")
	}
}

func TestErrorMetrics_Reset(t *testing.T) {
	m := NewErrorMetrics()
	m.RecordError(New(ErrBadRequest, "test"))
	m.RecordError(New(ErrNotFound, "test"))

	m.Reset()

	if m.totalErrors != 0 {
		t.Errorf("expected TotalErrors 0 after reset, got %d", m.totalErrors)
	}
	if len(m.errorCounts) != 0 {
		t.Errorf("expected empty ErrorCounts after reset, got %d", len(m.errorCounts))
	}
	if len(m.severityCounts) != 0 {
		t.Errorf("expected empty SeverityCounts after reset, got %d", len(m.severityCounts))
	}
	if len(m.categoryCounts) != 0 {
		t.Errorf("expected empty CategoryCounts after reset, got %d", len(m.categoryCounts))
	}
	if m.errorRate != 0 {
		t.Errorf("expected ErrorRate 0 after reset, got %f", m.errorRate)
	}
}

func TestNewErrorHookManager(t *testing.T) {
	mgr := NewErrorHookManager()
	if mgr == nil {
		t.Fatal("expected non-nil ErrorHookManager")
	}
	if mgr.hooks == nil {
		t.Error("expected initialized hooks slice")
	}
}

func TestErrorHookManager_AddHook(t *testing.T) {
	mgr := NewErrorHookManager()

	called := false
	mgr.AddHook(func(ctx context.Context, err *AppError) {
		called = true
	})

	if len(mgr.hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(mgr.hooks))
	}
	_ = called // hook is not executed yet
}

func TestErrorHookManager_ExecuteHooks(t *testing.T) {
	mgr := NewErrorHookManager()

	var mu sync.Mutex
	callCount := 0

	mgr.AddHook(func(ctx context.Context, err *AppError) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})
	mgr.AddHook(func(ctx context.Context, err *AppError) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	ctx := context.Background()
	appErr := New(ErrInternal, "test error")

	mgr.ExecuteHooks(ctx, appErr)

	// Hooks execute in goroutines, wait a bit
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if callCount != 2 {
		t.Errorf("expected 2 hook calls, got %d", callCount)
	}
	mu.Unlock()
}

func TestErrorHookManager_ExecuteHooks_PanicRecovery(t *testing.T) {
	mgr := NewErrorHookManager()

	var mu sync.Mutex
	secondCalled := false

	// First hook panics
	mgr.AddHook(func(ctx context.Context, err *AppError) {
		panic("hook panic")
	})

	// Second hook should still execute
	mgr.AddHook(func(ctx context.Context, err *AppError) {
		mu.Lock()
		secondCalled = true
		mu.Unlock()
	})

	ctx := context.Background()
	appErr := New(ErrInternal, "test")

	// Should not panic
	mgr.ExecuteHooks(ctx, appErr)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if !secondCalled {
		t.Error("expected second hook to execute despite first hook panic")
	}
	mu.Unlock()
}

func TestRecordErrorGlobal_NilError(t *testing.T) {
	// Should not panic
	RecordErrorGlobal(context.Background(), nil)
}

func TestRecordErrorGlobal_RecordsMetrics(t *testing.T) {
	ctx := context.Background()
	appErr := New(ErrBadRequest, "test")

	// Get metrics before
	metrics := GetGlobalMetrics()
	before := metrics.totalErrors

	RecordErrorGlobal(ctx, appErr)

	// Should have incremented
	if metrics.totalErrors != before+1 {
		t.Errorf("expected TotalErrors to increment by 1")
	}
}

func TestAlertingHook(t *testing.T) {
	var mu sync.Mutex
	alertedErrors := make([]*AppError, 0)

	hook := AlertingHook(func(ctx context.Context, err *AppError) {
		mu.Lock()
		alertedErrors = append(alertedErrors, err)
		mu.Unlock()
	})

	ctx := context.Background()

	// Critical error should trigger alert
	criticalErr := New(ErrRepoDatabase, "db down")
	hook(ctx, criticalErr)

	// High severity error should trigger alert
	highErr := New(ErrUnauthorized, "bad token")
	hook(ctx, highErr)

	// Low severity error should NOT trigger alert
	lowErr := New(ErrBadRequest, "bad input")
	hook(ctx, lowErr)

	mu.Lock()
	if len(alertedErrors) != 2 {
		t.Errorf("expected 2 alerted errors, got %d", len(alertedErrors))
	}
	mu.Unlock()
}

func TestMetricsHook(t *testing.T) {
	var mu sync.Mutex
	recorded := make([]string, 0)

	hook := MetricsHook(func(code string, severity ErrorSeverity, category ErrorCategory) {
		mu.Lock()
		recorded = append(recorded, code)
		mu.Unlock()
	})

	ctx := context.Background()
	hook(ctx, New(ErrBadRequest, "test"))
	hook(ctx, New(ErrNotFound, "test"))

	mu.Lock()
	if len(recorded) != 2 {
		t.Errorf("expected 2 recorded metrics, got %d", len(recorded))
	}
	if recorded[0] != ErrBadRequest {
		t.Errorf("expected first recorded code %s, got %s", ErrBadRequest, recorded[0])
	}
	mu.Unlock()
}
