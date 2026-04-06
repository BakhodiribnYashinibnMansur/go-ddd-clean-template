package query

import (
	"context"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestListHandler_Handle(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*integentity.IntegrationView{
			{ID: integentity.NewIntegrationID(), Name: "Slack", Type: "messaging", APIKey: "k1", Enabled: true, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: integentity.NewIntegrationID(), Name: "SMTP", Type: "email", APIKey: "k2", Enabled: false, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: integentity.IntegrationFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{views: []*integentity.IntegrationView{}, total: 0}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: integentity.IntegrationFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Integrations) != 0 {
		t.Errorf("expected 0 integrations, got %d", len(result.Integrations))
	}
}

func TestListHandler_WithFilters(t *testing.T) {
	t.Parallel()

	enabled := true
	intType := "messaging"
	readRepo := &mockReadRepo{
		views: []*integentity.IntegrationView{
			{ID: integentity.NewIntegrationID(), Name: "Slack", Type: "messaging", APIKey: "k", Enabled: true, Config: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: integentity.IntegrationFilter{Type: &intType, Enabled: &enabled, Limit: 10},
	})
	require.NoError(t, err)
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListQuery{Filter: integentity.IntegrationFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
