package command_test

import (
	"context"
	"testing"

	"gct/internal/context/ops/generic/metric/application/command"
	metricentity "gct/internal/context/ops/generic/metric/domain/entity"
	metricrepo "gct/internal/context/ops/generic/metric/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockMetricRepo struct {
	saved *metricentity.FunctionMetric
}

func (m *mockMetricRepo) Save(_ context.Context, fm *metricentity.FunctionMetric) error {
	m.saved = fm
	return nil
}

func (m *mockMetricRepo) List(_ context.Context, _ metricrepo.MetricFilter) ([]*metricentity.FunctionMetric, int64, error) {
	return nil, 0, nil
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                               {}
func (m *mockLogger) Debugf(_ string, _ ...any)                    {}
func (m *mockLogger) Debugw(_ string, _ ...any)                    {}
func (m *mockLogger) Info(_ ...any)                                {}
func (m *mockLogger) Infof(_ string, _ ...any)                     {}
func (m *mockLogger) Infow(_ string, _ ...any)                     {}
func (m *mockLogger) Warn(_ ...any)                                {}
func (m *mockLogger) Warnf(_ string, _ ...any)                     {}
func (m *mockLogger) Warnw(_ string, _ ...any)                     {}
func (m *mockLogger) Error(_ ...any)                               {}
func (m *mockLogger) Errorf(_ string, _ ...any)                    {}
func (m *mockLogger) Errorw(_ string, _ ...any)                    {}
func (m *mockLogger) Fatal(_ ...any)                               {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                    {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

func TestRecordMetricHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockMetricRepo{}
	handler := command.NewRecordMetricHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	cmd := command.RecordMetricCommand{
		Name:      "UserService.Create",
		LatencyMs: 150.5,
		IsPanic:   false,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected metric to be saved")
	}
	if repo.saved.Name() != "UserService.Create" {
		t.Fatalf("expected name UserService.Create, got %s", repo.saved.Name())
	}
	if repo.saved.LatencyMs() != 150.5 {
		t.Fatalf("expected latencyMs 150.5, got %f", repo.saved.LatencyMs())
	}
	if repo.saved.IsPanic() {
		t.Fatal("expected isPanic false")
	}
}
