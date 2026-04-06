package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"
	"time"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"

	"github.com/stretchr/testify/require"
)

func TestListIPRulesHandler_Handle(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*iprulerepo.IPRuleView{
			{ID: ipruleentity.NewIPRuleID(), IPAddress: "1.1.1.1", Action: "DENY", Reason: "r1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: ipruleentity.NewIPRuleID(), IPAddress: "2.2.2.2", Action: "ALLOW", Reason: "r2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListIPRulesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListIPRulesQuery{
		Filter: iprulerepo.IPRuleFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.IPRules) != 2 {
		t.Fatalf("expected 2 ip rules, got %d", len(result.IPRules))
	}
	if result.IPRules[0].IPAddress != "1.1.1.1" {
		t.Errorf("expected 1.1.1.1, got %s", result.IPRules[0].IPAddress)
	}
}

func TestListIPRulesHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{views: []*iprulerepo.IPRuleView{}, total: 0}

	handler := NewListIPRulesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListIPRulesQuery{
		Filter: iprulerepo.IPRuleFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.IPRules) != 0 {
		t.Errorf("expected 0 ip rules, got %d", len(result.IPRules))
	}
}

func TestListIPRulesHandler_WithFilters(t *testing.T) {
	t.Parallel()

	action := "DENY"
	readRepo := &mockReadRepo{
		views: []*iprulerepo.IPRuleView{
			{ID: ipruleentity.NewIPRuleID(), IPAddress: "3.3.3.3", Action: "DENY", Reason: "blocked", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListIPRulesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListIPRulesQuery{
		Filter: iprulerepo.IPRuleFilter{Action: &action, Limit: 10},
	})
	require.NoError(t, err)
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListIPRulesHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListIPRulesHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListIPRulesQuery{Filter: iprulerepo.IPRuleFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
