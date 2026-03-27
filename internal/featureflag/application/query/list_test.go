package query

import (
	"context"
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestListHandler_Handle(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.FeatureFlagView{
			{ID: uuid.New(), Name: "dark_mode", Description: "Dark mode", Enabled: true, RolloutPercentage: 100},
			{ID: uuid.New(), Name: "new_ui", Description: "New UI", Enabled: false, RolloutPercentage: 0},
		},
		total: 2,
	}

	handler := NewListHandler(readRepo)
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result.Flags))
	}
	if result.Flags[0].Name != "dark_mode" {
		t.Errorf("expected dark_mode, got %s", result.Flags[0].Name)
	}
}

func TestListHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.FeatureFlagView{}, total: 0}

	handler := NewListHandler(readRepo)
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.FeatureFlagFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Flags) != 0 {
		t.Errorf("expected 0 flags, got %d", len(result.Flags))
	}
}

func TestListHandler_WithFilters(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.FeatureFlagView{
			{ID: uuid.New(), Name: "dark_mode", Description: "Dark", Enabled: true, RolloutPercentage: 100},
		},
		total: 1,
	}

	handler := NewListHandler(readRepo)
	enabled := true

	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.FeatureFlagFilter{
			Enabled: &enabled,
			Limit:   10,
		},
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
	handler := NewListHandler(readRepo)
	_, err := handler.Handle(context.Background(), ListQuery{Filter: domain.FeatureFlagFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
