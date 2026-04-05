package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/ops/iprule/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.IPRuleView
	views []*domain.IPRuleView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.IPRuleView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrIPRuleNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.IPRuleFilter) ([]*domain.IPRuleView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.IPRuleView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.IPRuleFilter) ([]*domain.IPRuleView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetIPRuleHandler_Handle(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	readRepo := &mockReadRepo{
		view: &domain.IPRuleView{
			ID:        id,
			IPAddress: "192.168.1.100",
			Action:    "DENY",
			Reason:    "suspicious",
			ExpiresAt: &expires,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.IPAddress != "192.168.1.100" {
		t.Errorf("expected ip 192.168.1.100, got %s", result.IPAddress)
	}
	if result.Action != "DENY" {
		t.Errorf("expected action DENY, got %s", result.Action)
	}
	if result.ExpiresAt == nil {
		t.Error("expected expiresAt to be set")
	}
}

func TestGetIPRuleHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetIPRuleHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetIPRuleHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.IPRuleView{
			ID:        id,
			IPAddress: "10.0.0.1",
			Action:    "ALLOW",
			Reason:    "trusted",
			ExpiresAt: nil,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Reason != "trusted" {
		t.Errorf("expected reason 'trusted', got %s", result.Reason)
	}
	if result.ExpiresAt != nil {
		t.Error("expected expiresAt to be nil")
	}
}
