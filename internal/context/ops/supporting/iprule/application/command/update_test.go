package command

import (
	"context"
	"testing"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

func TestUpdateIPRuleHandler_Handle(t *testing.T) {
	t.Parallel()

	r := ipruleentity.NewIPRule("192.168.1.1", "DENY", "bad actor", nil)

	repo := &mockIPRuleRepo{
		findFn: func(_ context.Context, id ipruleentity.IPRuleID) (*ipruleentity.IPRule, error) {
			if id == r.TypedID() {
				return r, nil
			}
			return nil, ipruleentity.ErrIPRuleNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	newIP := "10.0.0.1"
	newAction := "ALLOW"
	newReason := "now trusted"
	cmd := UpdateIPRuleCommand{
		ID:        r.TypedID(),
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

	r := ipruleentity.NewIPRule("192.168.1.1", "DENY", "reason", nil)

	repo := &mockIPRuleRepo{
		findFn: func(_ context.Context, _ ipruleentity.IPRuleID) (*ipruleentity.IPRule, error) {
			return r, nil
		},
	}
	handler := NewUpdateIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	newReason := "updated reason"
	err := handler.Handle(context.Background(), UpdateIPRuleCommand{
		ID:     r.TypedID(),
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
	handler := NewUpdateIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), UpdateIPRuleCommand{ID: ipruleentity.NewIPRuleID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
