package domain

import (
	"time"

	"github.com/google/uuid"
)

// RoleCreated is raised when a new role is created.
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

// RoleDeleted is raised when a role is deleted.
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

// PolicyUpdated is raised when a policy is updated.
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

// PermissionGranted is raised when a permission is granted to a role.
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
