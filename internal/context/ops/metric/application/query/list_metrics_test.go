package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/metric/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	views []*domain.MetricView
	total int64
}

func (m *mockReadRepo) List(_ context.Context, _ domain.MetricFilter) ([]*domain.MetricView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) List(_ context.Context, _ domain.MetricFilter) ([]*domain.MetricView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: ListMetrics ---

func TestListMetricsHandler_Handle(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.MetricView{
			{ID: uuid.New(), Name: "UserService.Create", LatencyMs: 150.5, IsPanic: false, CreatedAt: now},
			{ID: uuid.New(), Name: "AuthService.Login", LatencyMs: 300.0, IsPanic: false, CreatedAt: now},
		},
		total: 2,
	}

	handler := NewListMetricsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(result.Metrics))
	}
	if result.Metrics[0].Name != "UserService.Create" {
		t.Errorf("expected 'UserService.Create', got %s", result.Metrics[0].Name)
	}
	if result.Metrics[0].LatencyMs != 150.5 {
		t.Errorf("expected latency 150.5, got %f", result.Metrics[0].LatencyMs)
	}
}

func TestListMetricsHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.MetricView{}, total: 0}

	handler := NewListMetricsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListMetricsQuery{
		Filter: domain.MetricFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Metrics) != 0 {
		t.Errorf("expected 0 metrics, got %d", len(result.Metrics))
	}
}

func TestListMetricsHandler_WithPanicError(t *testing.T) {
	panicErr := "nil pointer dereference"
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.MetricView{
			{ID: uuid.New(), Name: "Handler.Crash", LatencyMs: 10.0, IsPanic: true, PanicError: &panicErr, CreatedAt: now},
		},
		total: 1,
	}

	handler := NewListMetricsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if !result.Metrics[0].IsPanic {
		t.Error("expected IsPanic true")
	}
	if result.Metrics[0].PanicError == nil || *result.Metrics[0].PanicError != "nil pointer dereference" {
		t.Error("panic error not mapped correctly")
	}
}

func TestListMetricsHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListMetricsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListMetricsQuery{Filter: domain.MetricFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
