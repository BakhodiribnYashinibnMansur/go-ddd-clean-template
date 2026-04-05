package iprule

import (
	"context"
	"testing"

	"gct/internal/context/ops/iprule"
	"gct/internal/context/ops/iprule/application/command"
	"gct/internal/context/ops/iprule/application/query"
	"gct/internal/context/ops/iprule/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *iprule.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return iprule.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetIPRule(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIPRule.Handle(ctx, command.CreateIPRuleCommand{
		IPAddress: "192.168.1.100",
		Action:    "block",
		Reason:    "Suspicious activity",
		ExpiresAt: nil,
	})
	if err != nil {
		t.Fatalf("CreateIPRule: %v", err)
	}

	result, err := bc.ListIPRules.Handle(ctx, query.ListIPRulesQuery{
		Filter: domain.IPRuleFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListIPRules: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 ip rule, got %d", result.Total)
	}

	r := result.IPRules[0]
	if r.IPAddress != "192.168.1.100" {
		t.Errorf("expected ip_address 192.168.1.100, got %s", r.IPAddress)
	}
	if r.Action != "block" {
		t.Errorf("expected action block, got %s", r.Action)
	}

	view, err := bc.GetIPRule.Handle(ctx, query.GetIPRuleQuery{ID: r.ID})
	if err != nil {
		t.Fatalf("GetIPRule: %v", err)
	}
	if view.ID != r.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, r.ID)
	}
}

func TestIntegration_UpdateIPRule(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIPRule.Handle(ctx, command.CreateIPRuleCommand{
		IPAddress: "10.0.0.1",
		Action:    "allow",
		Reason:    "Trusted office IP",
		ExpiresAt: nil,
	})
	if err != nil {
		t.Fatalf("CreateIPRule: %v", err)
	}

	list, _ := bc.ListIPRules.Handle(ctx, query.ListIPRulesQuery{
		Filter: domain.IPRuleFilter{Limit: 10},
	})
	rID := list.IPRules[0].ID

	newAction := "block"
	newReason := "No longer trusted"
	err = bc.UpdateIPRule.Handle(ctx, command.UpdateIPRuleCommand{
		ID:     rID,
		Action: &newAction,
		Reason: &newReason,
	})
	if err != nil {
		t.Fatalf("UpdateIPRule: %v", err)
	}

	view, _ := bc.GetIPRule.Handle(ctx, query.GetIPRuleQuery{ID: rID})
	if view.Action != "block" {
		t.Errorf("action not updated, got %s", view.Action)
	}
	if view.Reason != "No longer trusted" {
		t.Errorf("reason not updated, got %s", view.Reason)
	}
}

func TestIntegration_DeleteIPRule(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIPRule.Handle(ctx, command.CreateIPRuleCommand{
		IPAddress: "172.16.0.1",
		Action:    "block",
		Reason:    "Temporary block",
		ExpiresAt: nil,
	})
	if err != nil {
		t.Fatalf("CreateIPRule: %v", err)
	}

	list, _ := bc.ListIPRules.Handle(ctx, query.ListIPRulesQuery{
		Filter: domain.IPRuleFilter{Limit: 10},
	})
	rID := list.IPRules[0].ID

	err = bc.DeleteIPRule.Handle(ctx, command.DeleteIPRuleCommand{ID: rID})
	if err != nil {
		t.Fatalf("DeleteIPRule: %v", err)
	}

	list2, _ := bc.ListIPRules.Handle(ctx, query.ListIPRulesQuery{
		Filter: domain.IPRuleFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 ip rules after delete, got %d", list2.Total)
	}
}
