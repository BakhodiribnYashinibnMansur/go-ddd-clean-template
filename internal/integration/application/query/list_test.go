package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/integration/domain"

	"github.com/google/uuid"
)

func TestListHandler_Handle(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.IntegrationView{
			{ID: uuid.New(), Name: "Slack", Type: "messaging", APIKey: "k1", Enabled: true, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), Name: "SMTP", Type: "email", APIKey: "k2", Enabled: false, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.IntegrationFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Integrations) != 2 {
		t.Fatalf("expected 2 integrations, got %d", len(result.Integrations))
	}
	if result.Integrations[0].Name != "Slack" {
		t.Errorf("expected Slack, got %s", result.Integrations[0].Name)
	}
}

func TestListHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.IntegrationView{}, total: 0}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.IntegrationFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Integrations) != 0 {
		t.Errorf("expected 0 integrations, got %d", len(result.Integrations))
	}
}

func TestListHandler_WithFilters(t *testing.T) {
	enabled := true
	intType := "messaging"
	readRepo := &mockReadRepo{
		views: []*domain.IntegrationView{
			{ID: uuid.New(), Name: "Slack", Type: "messaging", APIKey: "k", Enabled: true, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.IntegrationFilter{Type: &intType, Enabled: &enabled, Limit: 10},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListQuery{Filter: domain.IntegrationFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
