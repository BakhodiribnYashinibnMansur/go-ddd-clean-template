package metric

import (
	"context"
	"testing"

	"gct/internal/context/ops/generic/metric"
	"gct/internal/context/ops/generic/metric/application/command"
	"gct/internal/context/ops/generic/metric/application/query"
	metricrepo "gct/internal/context/ops/generic/metric/domain/repository"
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
		Filter: metricrepo.MetricFilter{Limit: 10},
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
		Filter: metricrepo.MetricFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("expected 3 metrics, got %d", result.Total)
	}
}

func TestIntegration_ListMetrics_FilterByName(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	commands := []command.RecordMetricCommand{
		{Name: "GetUser", LatencyMs: 10.0, IsPanic: false},
		{Name: "CreateUser", LatencyMs: 20.0, IsPanic: false},
		{Name: "GetUser", LatencyMs: 30.0, IsPanic: false},
	}

	for _, cmd := range commands {
		if err := bc.RecordMetric.Handle(ctx, cmd); err != nil {
			t.Fatalf("RecordMetric (%s): %v", cmd.Name, err)
		}
	}

	nameFilter := "GetUser"
	result, err := bc.ListMetrics.Handle(ctx, query.ListMetricsQuery{
		Filter: metricrepo.MetricFilter{
			Name:  &nameFilter,
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 metrics matching name filter, got %d", result.Total)
	}
	for _, m := range result.Metrics {
		if m.Name != "GetUser" {
			t.Errorf("expected all metrics to have name GetUser, got %s", m.Name)
		}
	}
}

func TestIntegration_ListMetrics_Pagination(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := bc.RecordMetric.Handle(ctx, command.RecordMetricCommand{
			Name:      "PaginatedOp",
			LatencyMs: float64(i + 1),
			IsPanic:   false,
		})
		if err != nil {
			t.Fatalf("RecordMetric #%d: %v", i, err)
		}
	}

	result, err := bc.ListMetrics.Handle(ctx, query.ListMetricsQuery{
		Filter: metricrepo.MetricFilter{
			Limit:  2,
			Offset: 0,
		},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 5 {
		t.Fatalf("expected total=5, got %d", result.Total)
	}
	if len(result.Metrics) != 2 {
		t.Fatalf("expected 2 metrics returned, got %d", len(result.Metrics))
	}
}

func TestIntegration_RecordPanicMetric(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	panicErr := "runtime: index out of range"
	err := bc.RecordMetric.Handle(ctx, command.RecordMetricCommand{
		Name:       "DangerousOp",
		LatencyMs:  100.0,
		IsPanic:    true,
		PanicError: &panicErr,
	})
	if err != nil {
		t.Fatalf("RecordMetric: %v", err)
	}

	// Also record a non-panic metric to verify filtering
	err = bc.RecordMetric.Handle(ctx, command.RecordMetricCommand{
		Name:      "SafeOp",
		LatencyMs: 5.0,
		IsPanic:   false,
	})
	if err != nil {
		t.Fatalf("RecordMetric: %v", err)
	}

	isPanic := true
	result, err := bc.ListMetrics.Handle(ctx, query.ListMetricsQuery{
		Filter: metricrepo.MetricFilter{
			IsPanic: &isPanic,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("ListMetrics: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 panic metric, got %d", result.Total)
	}

	m := result.Metrics[0]
	if m.Name != "DangerousOp" {
		t.Errorf("expected name DangerousOp, got %s", m.Name)
	}
	if !m.IsPanic {
		t.Errorf("expected IsPanic true, got false")
	}
	if m.PanicError == nil {
		t.Fatalf("expected PanicError to be set, got nil")
	}
	if *m.PanicError != panicErr {
		t.Errorf("expected PanicError %q, got %q", panicErr, *m.PanicError)
	}
}
