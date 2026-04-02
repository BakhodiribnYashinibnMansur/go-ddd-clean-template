package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestListHandler_Handle(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	flag1 := &domain.FeatureFlagView{
		ID:        uuid.New(),
		Name:      "flag-1",
		Key:       "flag_1",
		FlagType:  "bool",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	flag2 := &domain.FeatureFlagView{
		ID:        uuid.New(),
		Name:      "flag-2",
		Key:       "flag_2",
		FlagType:  "string",
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
			return []*domain.FeatureFlagView{flag1, flag2}, 2, nil
		},
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result.Flags))
	}
	if result.Flags[0].Name != "flag-1" {
		t.Errorf("expected first flag name flag-1, got %s", result.Flags[0].Name)
	}
	if result.Flags[1].Name != "flag-2" {
		t.Errorf("expected second flag name flag-2, got %s", result.Flags[1].Name)
	}
}

func TestListHandler_Handle_Empty(t *testing.T) {
	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
			return []*domain.FeatureFlagView{}, 0, nil
		},
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{})
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

func TestListHandler_Handle_WithFilter(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	var capturedFilter domain.FeatureFlagFilter

	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, filter domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
			capturedFilter = filter
			return []*domain.FeatureFlagView{
				{ID: uuid.New(), Name: "active-flag", Key: "active", FlagType: "bool", IsActive: true, CreatedAt: now, UpdatedAt: now},
			}, 1, nil
		},
	}

	handler := NewListHandler(readRepo, logger.Noop())
	search := "active"
	enabled := true
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.FeatureFlagFilter{
			Search:  &search,
			Enabled: &enabled,
			Limit:   20,
			Offset:  5,
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if capturedFilter.Search == nil || *capturedFilter.Search != "active" {
		t.Error("expected search filter to be passed through")
	}
	if capturedFilter.Enabled == nil || *capturedFilter.Enabled != true {
		t.Error("expected enabled filter to be passed through")
	}
	if capturedFilter.Limit != 20 {
		t.Errorf("expected limit 20, got %d", capturedFilter.Limit)
	}
	if capturedFilter.Offset != 5 {
		t.Errorf("expected offset 5, got %d", capturedFilter.Offset)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListHandler_Handle_RepoError(t *testing.T) {
	repoErr := errors.New("db failure")
	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
			return nil, 0, repoErr
		},
	}

	handler := NewListHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}
