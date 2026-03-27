package query

import (
	"context"
	"errors"
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.FeatureFlagView
	views []*domain.FeatureFlagView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.FeatureFlagView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetFeatureFlag ---

func TestGetHandler_Handle(t *testing.T) {
	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.FeatureFlagView{
			ID:                id,
			Name:              "dark_mode",
			Description:       "Enable dark mode",
			Enabled:           true,
			RolloutPercentage: 50,
		},
	}

	handler := NewGetHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Name != "dark_mode" {
		t.Errorf("expected name dark_mode, got %s", result.Name)
	}
	if result.Description != "Enable dark mode" {
		t.Errorf("expected description, got %s", result.Description)
	}
	if result.Enabled != true {
		t.Errorf("expected enabled true, got %v", result.Enabled)
	}
	if result.RolloutPercentage != 50 {
		t.Errorf("expected rollout 50, got %d", result.RolloutPercentage)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()

	readRepo := &mockReadRepo{
		view: &domain.FeatureFlagView{
			ID:                id,
			Name:              "beta_feature",
			Description:       "Beta only",
			Enabled:           false,
			RolloutPercentage: 25,
		},
	}

	handler := NewGetHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != id {
		t.Error("ID not mapped correctly")
	}
	if result.Name != "beta_feature" {
		t.Error("name not mapped")
	}
	if result.Description != "Beta only" {
		t.Error("description not mapped")
	}
	if result.Enabled != false {
		t.Error("enabled not mapped")
	}
	if result.RolloutPercentage != 25 {
		t.Errorf("rollout percentage not mapped, got %d", result.RolloutPercentage)
	}
}
