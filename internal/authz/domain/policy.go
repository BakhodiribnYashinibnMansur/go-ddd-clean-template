package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// PolicyEffect represents the effect of a policy.
type PolicyEffect string

const (
	PolicyAllow PolicyEffect = "ALLOW"
	PolicyDeny  PolicyEffect = "DENY"
)

// Policy is an ABAC policy entity.
type Policy struct {
	shared.BaseEntity
	permissionID uuid.UUID
	effect       PolicyEffect
	priority     int
	active       bool
	conditions   map[string]any
}

// NewPolicy creates a new Policy with a generated ID.
func NewPolicy(permissionID uuid.UUID, effect PolicyEffect) *Policy {
	return &Policy{
		BaseEntity:   shared.NewBaseEntity(),
		permissionID: permissionID,
		effect:       effect,
		priority:     0,
		active:       true,
		conditions:   make(map[string]any),
	}
}

// ReconstructPolicy rebuilds a Policy from persisted data.
func ReconstructPolicy(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	permissionID uuid.UUID,
	effect PolicyEffect,
	priority int,
	active bool,
	conditions map[string]any,
) *Policy {
	if conditions == nil {
		conditions = make(map[string]any)
	}
	return &Policy{
		BaseEntity:   shared.NewBaseEntityWithID(id, createdAt, updatedAt, deletedAt),
		permissionID: permissionID,
		effect:       effect,
		priority:     priority,
		active:       active,
		conditions:   conditions,
	}
}

// PermissionID returns the policy's permission ID.
func (p *Policy) PermissionID() uuid.UUID { return p.permissionID }

// Effect returns the policy effect.
func (p *Policy) Effect() PolicyEffect { return p.effect }

// Priority returns the policy priority.
func (p *Policy) Priority() int { return p.priority }

// IsActive returns whether the policy is active.
func (p *Policy) IsActive() bool { return p.active }

// Conditions returns the ABAC conditions.
func (p *Policy) Conditions() map[string]any { return p.conditions }

// Toggle flips the active state.
func (p *Policy) Toggle() {
	p.active = !p.active
	p.Touch()
}

// SetPriority sets the policy priority.
func (p *Policy) SetPriority(priority int) {
	p.priority = priority
	p.Touch()
}

// SetEffect sets the policy effect.
func (p *Policy) SetEffect(effect PolicyEffect) {
	p.effect = effect
	p.Touch()
}

// SetConditions sets the ABAC conditions.
func (p *Policy) SetConditions(conditions map[string]any) {
	if conditions == nil {
		conditions = make(map[string]any)
	}
	p.conditions = conditions
	p.Touch()
}
