package events

import (
	"github.com/google/uuid"
)

// Event names for the user BC's Published Language. Consumers import these
// constants; never hardcode the string literal at call sites.
const (
	EventUserCreatedV1         = "user.created.v1"
	EventUserSignedInV1        = "user.signed_in.v1"
	EventUserDeactivatedV1     = "user.deactivated.v1"
	EventUserPasswordChangedV1 = "user.password_changed.v1"
	EventUserApprovedV1        = "user.approved.v1"
	EventUserRoleChangedV1     = "user.role_changed.v1"
)

// UserCreatedV1 is published when a new user aggregate is persisted.
type UserCreatedV1 struct {
	BaseEvent
	Phone string `json:"phone"`
}

// NewUserCreatedV1 constructs the event with a fresh envelope.
func NewUserCreatedV1(userID uuid.UUID, phone string) UserCreatedV1 {
	return UserCreatedV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserCreatedV1, userID, 1)},
		Phone:     phone,
	}
}

// UserSignedInV1 is published after successful credential verification and
// session creation.
type UserSignedInV1 struct {
	BaseEvent
	SessionID uuid.UUID `json:"session_id"`
	IPAddress string    `json:"ip_address"`
}

// NewUserSignedInV1 constructs the event.
func NewUserSignedInV1(userID, sessionID uuid.UUID, ip string) UserSignedInV1 {
	return UserSignedInV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserSignedInV1, userID, 1)},
		SessionID: sessionID,
		IPAddress: ip,
	}
}

// UserDeactivatedV1 is published when an admin deactivates a user account.
type UserDeactivatedV1 struct {
	BaseEvent
}

// NewUserDeactivatedV1 constructs the event.
func NewUserDeactivatedV1(userID uuid.UUID) UserDeactivatedV1 {
	return UserDeactivatedV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserDeactivatedV1, userID, 1)},
	}
}

// UserPasswordChangedV1 is published after a successful password change.
type UserPasswordChangedV1 struct {
	BaseEvent
}

// NewUserPasswordChangedV1 constructs the event.
func NewUserPasswordChangedV1(userID uuid.UUID) UserPasswordChangedV1 {
	return UserPasswordChangedV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserPasswordChangedV1, userID, 1)},
	}
}

// UserApprovedV1 is published when a user's account is approved.
type UserApprovedV1 struct {
	BaseEvent
}

// NewUserApprovedV1 constructs the event.
func NewUserApprovedV1(userID uuid.UUID) UserApprovedV1 {
	return UserApprovedV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserApprovedV1, userID, 1)},
	}
}

// UserRoleChangedV1 is published when a user's role is reassigned. Carries
// both the previous (nullable) and new role IDs so subscribers can detect
// privilege escalation in audit trails.
type UserRoleChangedV1 struct {
	BaseEvent
	OldRoleID *uuid.UUID `json:"old_role_id,omitempty"`
	NewRoleID uuid.UUID  `json:"new_role_id"`
}

// NewUserRoleChangedV1 constructs the event.
func NewUserRoleChangedV1(userID uuid.UUID, oldRoleID *uuid.UUID, newRoleID uuid.UUID) UserRoleChangedV1 {
	return UserRoleChangedV1{
		BaseEvent: BaseEvent{Envelope: NewEnvelope(EventUserRoleChangedV1, userID, 1)},
		OldRoleID: oldRoleID,
		NewRoleID: newRoleID,
	}
}
