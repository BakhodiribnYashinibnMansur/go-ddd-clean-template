package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPolicy(t *testing.T) {
	t.Parallel()

	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)

	if policy.PermissionID() != permID {
		t.Errorf("expected permission ID %s, got %s", permID, policy.PermissionID())
	}

	if policy.Effect() != PolicyAllow {
		t.Errorf("expected effect ALLOW, got %s", policy.Effect())
	}

	if !policy.IsActive() {
		t.Error("expected policy to be active by default")
	}

	if policy.Priority() != 0 {
		t.Errorf("expected priority 0, got %d", policy.Priority())
	}
}

func TestPolicy_Toggle(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(uuid.New(), PolicyAllow)

	if !policy.IsActive() {
		t.Fatal("expected active")
	}

	policy.Toggle()
	if policy.IsActive() {
		t.Error("expected inactive after toggle")
	}

	policy.Toggle()
	if !policy.IsActive() {
		t.Error("expected active after second toggle")
	}
}

func TestPolicy_SetPriority(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(uuid.New(), PolicyDeny)
	policy.SetPriority(10)

	if policy.Priority() != 10 {
		t.Errorf("expected priority 10, got %d", policy.Priority())
	}
}

func TestPolicy_SetEffect(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(uuid.New(), PolicyAllow)
	policy.SetEffect(PolicyDeny)

	if policy.Effect() != PolicyDeny {
		t.Errorf("expected effect DENY, got %s", policy.Effect())
	}
}

func TestPolicy_SetConditions(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(uuid.New(), PolicyAllow)
	conditions := map[string]any{"ip_range": "10.0.0.0/8"}
	policy.SetConditions(conditions)

	if policy.Conditions()["ip_range"] != "10.0.0.0/8" {
		t.Errorf("expected ip_range condition, got %v", policy.Conditions())
	}
}

func TestPolicy_SetConditions_Nil(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(uuid.New(), PolicyAllow)
	policy.SetConditions(nil)

	if policy.Conditions() == nil {
		t.Error("expected non-nil conditions map after setting nil")
	}
}
