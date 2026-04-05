package metric

import (
	"context"
	"testing"

	"gct/internal/context/ops/metric"
	"gct/internal/context/ops/metric/application/command"
	"gct/internal/context/ops/metric/application/query"
	"gct/internal/context/ops/metric/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *metric.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return metric.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_RecordAndListMetrics(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.RecordMetric.Handle(ctx, command.RecordMetricCommand{
		Name:      "GetUser",
		LatencyMs: 42.5,
		IsPanic:   false,
	})
	if err != nil {
		t.Fatalf("RecordMetric: %v", err)
	}

	result, err := bc.ListMetrics.Handle(ctx, query.ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 metric, got %d", result.Total)
	}

	m := result.Metrics[0]
	if m.Name != "GetUser" {
		t.Errorf("expected name GetUser, got %s", m.Name)
	}
	if m.LatencyMs != 42.0 {
		t.Errorf("expected latency 42.0, got %f", m.LatencyMs)
	}
	if m.IsPanic {
		t.Errorf("expected IsPanic false, got true")
	}
}

func TestIntegration_MultipleRecords(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	panicErr := "nil pointer dereference"
	metrics := []command.RecordMetricCommand{
		{Name: "CreateUser", LatencyMs: 15.2, IsPanic: false},
		{Name: "DeleteUser", LatencyMs: 8.1, IsPanic: false},
		{Name: "UpdateUser", LatencyMs: 500.0, IsPanic: true, PanicError: &panicErr},
	}

	for _, cmd := range metrics {
		if err := bc.RecordMetric.Handle(ctx, cmd); err != nil {
			t.Fatalf("RecordMetric (%s): %v", cmd.Name, err)
		}
	}

	result, err := bc.ListMetrics.Handle(ctx, query.ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("expected 3 metrics, got %d", result.Total)
	}
}
