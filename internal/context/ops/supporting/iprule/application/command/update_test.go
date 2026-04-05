package command

import (
	"context"
	"testing"

	"gct/internal/context/ops/supporting/iprule/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateIPRuleHandler_Handle(t *testing.T) {
	t.Parallel()

	r := domain.NewIPRule("192.168.1.1", "DENY", "bad actor", nil)

	repo := &mockIPRuleRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.IPRule, error) {
			if id == r.ID() {
				return r, nil
			}
			return nil, domain.ErrIPRuleNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateIPRuleHandler(repo, eb, &mockLogger{})

	newIP := "10.0.0.1"
	newAction := "ALLOW"
	newReason := "now trusted"
	cmd := UpdateIPRuleCommand{
		ID:        domain.IPRuleID(r.ID()),
		IPAddress: &newIP,
		Action:    &newAction,
		Reason:    &newReason,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updated == nil {
		t.Fatal("expected ip rule to be updated")
	}
	if repo.updated.IPAddress() != "10.0.0.1" {
		t.Errorf("expected ip 10.0.0.1, got %s", repo.updated.IPAddress())
	}
	if repo.updated.Action() != "ALLOW" {
		t.Errorf("expected action ALLOW, got %s", repo.updated.Action())
	}
	if repo.updated.Reason() != "now trusted" {
		t.Errorf("expected reason 'now trusted', got %s", repo.updated.Reason())
	}
}

func TestUpdateIPRuleHandler_PartialUpdate(t *testing.T) {
	t.Parallel()

	r := domain.NewIPRule("192.168.1.1", "DENY", "reason", nil)

	repo := &mockIPRuleRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.IPRule, error) {
			return r, nil
		},
	}
	handler := NewUpdateIPRuleHandler(repo, &mockEventBus{}, &mockLogger{})

	newReason := "updated reason"
	err := handler.Handle(context.Background(), UpdateIPRuleCommand{
		ID:     domain.IPRuleID(r.ID()),
		Reason: &newReason,
	})
	require.NoError(t, err)
	if repo.updated.IPAddress() != "192.168.1.1" {
		t.Error("ip should remain unchanged")
	}
	if repo.updated.Action() != "DENY" {
		t.Error("action should remain unchanged")
	}
	if repo.updated.Reason() != "updated reason" {
		t.Errorf("expected updated reason, got %s", repo.updated.Reason())
	}
}

func TestUpdateIPRuleHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockIPRuleRepo{}
	handler := NewUpdateIPRuleHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateIPRuleCommand{ID: domain.NewIPRuleID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
