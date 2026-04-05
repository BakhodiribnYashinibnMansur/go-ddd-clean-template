package errors_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	apperrors "gct/internal/platform/infrastructure/errors"
)

type mockEnqueuer struct {
	mu    sync.Mutex
	tasks []string
}

func (m *mockEnqueuer) EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := json.Marshal(payload)
	m.tasks = append(m.tasks, taskType+":"+string(b))
	return nil, nil
}

func TestAlerter_SendsCriticalErrors(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityCritical,
		DebouncePeriod: 0,
	}, nil, nil)

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	alerter.SendError(err)

	enq.mu.Lock()
	defer enq.mu.Unlock()
	if len(enq.tasks) != 1 {
		t.Fatalf("expected 1 task enqueued, got %d", len(enq.tasks))
	}
}

func TestAlerter_SkipsLowSeverity(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityHigh,
		DebouncePeriod: 0,
	}, nil, nil)

	err := apperrors.New(apperrors.ErrBadRequest, "")
	alerter.SendError(err)

	enq.mu.Lock()
	defer enq.mu.Unlock()
	if len(enq.tasks) != 0 {
		t.Fatalf("expected 0 tasks for low severity, got %d", len(enq.tasks))
	}
}

func TestAlerter_DebouncesSameCode(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityCritical,
		DebouncePeriod: 100 * time.Millisecond,
	}, nil, nil)

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	alerter.SendError(err)
	alerter.SendError(err)
	alerter.SendError(err)

	enq.mu.Lock()
	count := len(enq.tasks)
	enq.mu.Unlock()
	if count != 1 {
		t.Fatalf("expected 1 task (debounced), got %d", count)
	}
}
