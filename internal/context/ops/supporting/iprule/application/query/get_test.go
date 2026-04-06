package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *iprulerepo.IPRuleView
	views []*iprulerepo.IPRuleView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id ipruleentity.IPRuleID) (*iprulerepo.IPRuleView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, ipruleentity.ErrIPRuleNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ iprulerepo.IPRuleFilter) ([]*iprulerepo.IPRuleView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ ipruleentity.IPRuleID) (*iprulerepo.IPRuleView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ iprulerepo.IPRuleFilter) ([]*iprulerepo.IPRuleView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetIPRuleHandler_Handle(t *testing.T) {
	t.Parallel()

	id := ipruleentity.NewIPRuleID()
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	readRepo := &mockReadRepo{
		view: &iprulerepo.IPRuleView{
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
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: ipruleentity.NewIPRuleID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetIPRuleHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetIPRuleHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetIPRuleQuery{ID: ipruleentity.NewIPRuleID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetIPRuleHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := ipruleentity.NewIPRuleID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &iprulerepo.IPRuleView{
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
	require.NoError(t, err)
	if result.Reason != "trusted" {
		t.Errorf("expected reason 'trusted', got %s", result.Reason)
	}
	if result.ExpiresAt != nil {
		t.Error("expected expiresAt to be nil")
	}
}
