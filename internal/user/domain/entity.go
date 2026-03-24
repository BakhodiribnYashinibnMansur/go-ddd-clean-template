package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

const maxSessions = 10

// ---------------------------------------------------------------------------
// User Aggregate Root
// ---------------------------------------------------------------------------

type User struct {
	shared.AggregateRoot
	phone      Phone
	email      *Email
	username   *string
	password   Password
	roleID     *uuid.UUID
	attributes map[string]any
	active     bool
	isApproved bool
	lastSeen   *time.Time
	sessions   []Session
}

// ---------------------------------------------------------------------------
// Functional Options
// ---------------------------------------------------------------------------

type UserOption func(*User)

func WithEmail(email Email) UserOption       { return func(u *User) { u.email = &email } }
func WithUsername(name string) UserOption     { return func(u *User) { u.username = &name } }
func WithRoleID(id uuid.UUID) UserOption     { return func(u *User) { u.roleID = &id } }
func WithAttributes(attrs map[string]any) UserOption {
	return func(u *User) { u.attributes = attrs }
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NewUser creates a brand-new User aggregate. It is active but not yet approved.
func NewUser(phone Phone, password Password, opts ...UserOption) *User {
	u := &User{
		AggregateRoot: shared.NewAggregateRoot(),
		phone:         phone,
		password:      password,
		attributes:    make(map[string]any),
		active:        true,
		isApproved:    false,
		sessions:      make([]Session, 0),
	}
	for _, opt := range opts {
		opt(u)
	}
	u.AddEvent(NewUserCreated(u.ID(), phone.Value()))
	return u
}

// ReconstructUser rebuilds a User aggregate from persisted data. No events are raised.
func ReconstructUser(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	phone Phone,
	email *Email,
	username *string,
	password Password,
	roleID *uuid.UUID,
	attributes map[string]any,
	active, isApproved bool,
	lastSeen *time.Time,
	sessions []Session,
) *User {
	if attributes == nil {
		attributes = make(map[string]any)
	}
	if sessions == nil {
		sessions = make([]Session, 0)
	}
	return &User{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		phone:         phone,
		email:         email,
		username:      username,
		password:      password,
		roleID:        roleID,
		attributes:    attributes,
		active:        active,
		isApproved:    isApproved,
		lastSeen:      lastSeen,
		sessions:      sessions,
	}
}

// ---------------------------------------------------------------------------
// Domain behaviour
// ---------------------------------------------------------------------------

// AddSession creates a new session and appends it to the aggregate.
// Returns an error if the user already has the maximum number of sessions.
func (u *User) AddSession(deviceType SessionDeviceType, ip, userAgent string) (*Session, error) {
	if len(u.sessions) >= maxSessions {
		return nil, ErrMaxSessionsReached
	}
	s := NewSession(u.ID(), deviceType, ip, userAgent)
	u.sessions = append(u.sessions, *s)
	u.Touch()
	u.AddEvent(NewUserSignedIn(u.ID(), s.ID(), ip))
	return s, nil
}

// RemoveSession removes a session by ID.
func (u *User) RemoveSession(sessionID uuid.UUID) error {
	for i, s := range u.sessions {
		if s.ID() == sessionID {
			u.sessions = append(u.sessions[:i], u.sessions[i+1:]...)
			u.Touch()
			return nil
		}
	}
	return ErrSessionNotFound
}

// RevokeAllSessions revokes every session in the aggregate.
func (u *User) RevokeAllSessions() {
	for i := range u.sessions {
		u.sessions[i].Revoke()
	}
	u.Touch()
}

// VerifyPassword checks the raw password against the stored hash.
func (u *User) VerifyPassword(raw string) error {
	return u.password.Compare(raw)
}

// ChangePassword verifies the old password and replaces it with the new one.
func (u *User) ChangePassword(oldRaw, newRaw string) error {
	if err := u.password.Compare(oldRaw); err != nil {
		return err
	}
	pw, err := NewPasswordFromRaw(newRaw)
	if err != nil {
		return err
	}
	u.password = pw
	u.Touch()
	u.AddEvent(NewPasswordChanged(u.ID()))
	return nil
}

// Activate marks the user as active.
func (u *User) Activate() {
	u.active = true
	u.Touch()
}

// Deactivate marks the user as inactive and raises a UserDeactivated event.
func (u *User) Deactivate() {
	u.active = false
	u.Touch()
	u.AddEvent(NewUserDeactivated(u.ID()))
}

// Approve marks the user as approved and raises a UserApproved event.
func (u *User) Approve() {
	u.isApproved = true
	u.Touch()
	u.AddEvent(NewUserApproved(u.ID()))
}

// ChangeRole sets a new role and raises a RoleChanged event.
func (u *User) ChangeRole(roleID uuid.UUID) {
	old := u.roleID
	u.roleID = &roleID
	u.Touch()
	u.AddEvent(NewRoleChanged(u.ID(), old, roleID))
}

// UpdateLastSeen sets lastSeen to now.
func (u *User) UpdateLastSeen() {
	now := time.Now()
	u.lastSeen = &now
	u.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (u *User) Phone() Phone           { return u.phone }
func (u *User) Email() *Email          { return u.email }
func (u *User) Username() *string      { return u.username }
func (u *User) Password() Password     { return u.password }
func (u *User) RoleID() *uuid.UUID     { return u.roleID }
func (u *User) Attributes() map[string]any { return u.attributes }
func (u *User) IsActive() bool         { return u.active }
func (u *User) IsApproved() bool       { return u.isApproved }
func (u *User) LastSeen() *time.Time   { return u.lastSeen }
func (u *User) Sessions() []Session    { return u.sessions }
