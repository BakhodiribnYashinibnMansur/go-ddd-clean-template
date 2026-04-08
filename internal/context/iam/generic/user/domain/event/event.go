package event

import (
	"time"

	"github.com/google/uuid"
)

// UserCreated is raised when a new user aggregate is instantiated via NewUser.
// Carries the phone number so downstream handlers can trigger welcome SMS without re-querying.
type UserCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Phone       string
}

func NewUserCreated(userID uuid.UUID, phone string) UserCreated {
	return UserCreated{
		aggregateID: userID,
		occurredAt:  time.Now(),
		Phone:       phone,
	}
}

func (e UserCreated) EventName() string      { return "user.created" }
func (e UserCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e UserCreated) AggregateID() uuid.UUID { return e.aggregateID }

// UserSignedIn is raised after successful credential verification and session creation.
// Carries session ID and IP for audit logging and anomaly detection (e.g., new-IP alerts).
type UserSignedIn struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	SessionID   uuid.UUID
	IPAddress   string
}

func NewUserSignedIn(userID, sessionID uuid.UUID, ip string) UserSignedIn {
	return UserSignedIn{
		aggregateID: userID,
		occurredAt:  time.Now(),
		SessionID:   sessionID,
		IPAddress:   ip,
	}
}

func (e UserSignedIn) EventName() string      { return "user.signed_in" }
func (e UserSignedIn) OccurredAt() time.Time  { return e.occurredAt }
func (e UserSignedIn) AggregateID() uuid.UUID { return e.aggregateID }

// UserDeactivated is raised when an admin deactivates a user account.
// Subscribers should consider revoking active sessions or sending a notification.
type UserDeactivated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewUserDeactivated(userID uuid.UUID) UserDeactivated {
	return UserDeactivated{
		aggregateID: userID,
		occurredAt:  time.Now(),
	}
}

func (e UserDeactivated) EventName() string      { return "user.deactivated" }
func (e UserDeactivated) OccurredAt() time.Time  { return e.occurredAt }
func (e UserDeactivated) AggregateID() uuid.UUID { return e.aggregateID }

// PasswordChanged is raised after a successful password change.
// Subscribers should invalidate all refresh tokens or notify the user of the change.
type PasswordChanged struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewPasswordChanged(userID uuid.UUID) PasswordChanged {
	return PasswordChanged{
		aggregateID: userID,
		occurredAt:  time.Now(),
	}
}

func (e PasswordChanged) EventName() string      { return "user.password_changed" }
func (e PasswordChanged) OccurredAt() time.Time  { return e.occurredAt }
func (e PasswordChanged) AggregateID() uuid.UUID { return e.aggregateID }

// UserApproved is raised when a user is approved.
type UserApproved struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewUserApproved(userID uuid.UUID) UserApproved {
	return UserApproved{
		aggregateID: userID,
		occurredAt:  time.Now(),
	}
}

func (e UserApproved) EventName() string      { return "user.approved" }
func (e UserApproved) OccurredAt() time.Time  { return e.occurredAt }
func (e UserApproved) AggregateID() uuid.UUID { return e.aggregateID }

// UserProfileUpdated is raised when a user's profile fields (email, username, attributes) are modified.
type UserProfileUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewUserProfileUpdated(userID uuid.UUID) UserProfileUpdated {
	return UserProfileUpdated{
		aggregateID: userID,
		occurredAt:  time.Now(),
	}
}

func (e UserProfileUpdated) EventName() string      { return "user.profile_updated" }
func (e UserProfileUpdated) OccurredAt() time.Time  { return e.occurredAt }
func (e UserProfileUpdated) AggregateID() uuid.UUID { return e.aggregateID }

// RoleChanged is raised when a user's role is changed. Carries both old (nullable, for first assignment)
// and new role IDs so subscribers can detect privilege escalation in audit logs.
type RoleChanged struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	OldRoleID   *uuid.UUID
	NewRoleID   uuid.UUID
}

func NewRoleChanged(userID uuid.UUID, oldRoleID *uuid.UUID, newRoleID uuid.UUID) RoleChanged {
	return RoleChanged{
		aggregateID: userID,
		occurredAt:  time.Now(),
		OldRoleID:   oldRoleID,
		NewRoleID:   newRoleID,
	}
}

func (e RoleChanged) EventName() string      { return "user.role_changed" }
func (e RoleChanged) OccurredAt() time.Time  { return e.occurredAt }
func (e RoleChanged) AggregateID() uuid.UUID { return e.aggregateID }

// ---------------------------------------------------------------------------
// V2 events — carry field-level changes for activity logging
// ---------------------------------------------------------------------------

// FieldChange represents a single field-level mutation (local copy of contract type
// to avoid the domain importing the contract package).
type FieldChange struct {
	FieldName string
	OldValue  string
	NewValue  string
}

// UserCreatedWithChanges is raised alongside UserCreated to carry initial field values.
type UserCreatedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
	Changes     []FieldChange
}

func NewUserCreatedWithChanges(userID, actorID uuid.UUID, changes []FieldChange) UserCreatedWithChanges {
	return UserCreatedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
		Changes:     changes,
	}
}

func (e UserCreatedWithChanges) EventName() string      { return "user.created.v2" }
func (e UserCreatedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e UserCreatedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }

// UserProfileUpdatedWithChanges carries field-level diffs for profile updates.
type UserProfileUpdatedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
	Changes     []FieldChange
}

func NewUserProfileUpdatedWithChanges(userID, actorID uuid.UUID, changes []FieldChange) UserProfileUpdatedWithChanges {
	return UserProfileUpdatedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
		Changes:     changes,
	}
}

func (e UserProfileUpdatedWithChanges) EventName() string      { return "user.profile_updated.v2" }
func (e UserProfileUpdatedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e UserProfileUpdatedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }

// UserDeletedWithChanges records user deletion with actor identity.
type UserDeletedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
}

func NewUserDeletedWithChanges(userID, actorID uuid.UUID) UserDeletedWithChanges {
	return UserDeletedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
	}
}

func (e UserDeletedWithChanges) EventName() string      { return "user.deleted.v2" }
func (e UserDeletedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e UserDeletedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }

// RoleChangedWithChanges carries the role change as field-level diff.
type RoleChangedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
	Changes     []FieldChange
}

func NewRoleChangedWithChanges(userID, actorID uuid.UUID, changes []FieldChange) RoleChangedWithChanges {
	return RoleChangedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
		Changes:     changes,
	}
}

func (e RoleChangedWithChanges) EventName() string      { return "user.role_changed.v2" }
func (e RoleChangedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e RoleChangedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }

// UserApprovedWithChanges carries the approval state change.
type UserApprovedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
	Changes     []FieldChange
}

func NewUserApprovedWithChanges(userID, actorID uuid.UUID, changes []FieldChange) UserApprovedWithChanges {
	return UserApprovedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
		Changes:     changes,
	}
}

func (e UserApprovedWithChanges) EventName() string      { return "user.approved.v2" }
func (e UserApprovedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e UserApprovedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }

// PasswordChangedWithChanges records a password change with actor identity.
type PasswordChangedWithChanges struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ActorID     uuid.UUID
	Changes     []FieldChange
}

func NewPasswordChangedWithChanges(userID, actorID uuid.UUID, changes []FieldChange) PasswordChangedWithChanges {
	return PasswordChangedWithChanges{
		aggregateID: userID,
		occurredAt:  time.Now(),
		ActorID:     actorID,
		Changes:     changes,
	}
}

func (e PasswordChangedWithChanges) EventName() string      { return "user.password_changed.v2" }
func (e PasswordChangedWithChanges) OccurredAt() time.Time  { return e.occurredAt }
func (e PasswordChangedWithChanges) AggregateID() uuid.UUID { return e.aggregateID }
