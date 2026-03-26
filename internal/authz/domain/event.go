package domain

import (
	"time"

	"github.com/google/uuid"
)

// RoleCreated is a domain event emitted when a new authorization role is instantiated.
// Consumers may use this to seed default permissions or notify admin dashboards.
type RoleCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Name        string
}

func NewRoleCreated(roleID uuid.UUID, name string) RoleCreated {
	return RoleCreated{
		aggregateID: roleID,
		occurredAt:  time.Now(),
		Name:        name,
	}
}

func (e RoleCreated) EventName() string      { return "authz.role_created" }
func (e RoleCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e RoleCreated) AggregateID() uuid.UUID { return e.aggregateID }

// RoleDeleted is emitted when a role is permanently removed.
// Consumers should cascade-revoke any user-role assignments referencing this role.
type RoleDeleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewRoleDeleted(roleID uuid.UUID) RoleDeleted {
	return RoleDeleted{
		aggregateID: roleID,
		occurredAt:  time.Now(),
	}
}

func (e RoleDeleted) EventName() string      { return "authz.role_deleted" }
func (e RoleDeleted) OccurredAt() time.Time  { return e.occurredAt }
func (e RoleDeleted) AggregateID() uuid.UUID { return e.aggregateID }

// PolicyUpdated is emitted when an ABAC policy's effect, priority, or conditions change.
// The aggregateID is the owning role, not the policy itself — enabling role-level event streams.
type PolicyUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	PolicyID    uuid.UUID
}

func NewPolicyUpdated(roleID, policyID uuid.UUID) PolicyUpdated {
	return PolicyUpdated{
		aggregateID: roleID,
		occurredAt:  time.Now(),
		PolicyID:    policyID,
	}
}

func (e PolicyUpdated) EventName() string      { return "authz.policy_updated" }
func (e PolicyUpdated) OccurredAt() time.Time  { return e.occurredAt }
func (e PolicyUpdated) AggregateID() uuid.UUID { return e.aggregateID }

// PermissionGranted is emitted when a permission is added to a role's permission set.
// This can trigger permission cache invalidation for all users assigned to the role.
type PermissionGranted struct {
	aggregateID  uuid.UUID
	occurredAt   time.Time
	PermissionID uuid.UUID
}

func NewPermissionGranted(roleID, permissionID uuid.UUID) PermissionGranted {
	return PermissionGranted{
		aggregateID:  roleID,
		occurredAt:   time.Now(),
		PermissionID: permissionID,
	}
}

func (e PermissionGranted) EventName() string      { return "authz.permission_granted" }
func (e PermissionGranted) OccurredAt() time.Time  { return e.occurredAt }
func (e PermissionGranted) AggregateID() uuid.UUID { return e.aggregateID }
