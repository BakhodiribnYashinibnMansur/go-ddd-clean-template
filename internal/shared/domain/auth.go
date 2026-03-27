package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuthSession is a cross-BC read model used by middleware to carry
// authenticated session identity through the request context.
// Both User BC (auth middleware) and Authz BC (authz middleware) depend on this type.
type AuthSession struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	DeviceID         uuid.UUID `json:"device_id"`
	RefreshTokenHash string    `json:"-"`
	ExpiresAt        time.Time `json:"expires_at"`
	Revoked          bool      `json:"revoked"`
	LastActivity     time.Time `json:"last_activity"`
}

// IsExpired reports whether the session has passed its expiration time.
func (s *AuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive reports whether the session is usable (not expired and not revoked).
func (s *AuthSession) IsActive() bool {
	return !s.IsExpired() && !s.Revoked
}

// AuthUser is a cross-BC read model carrying minimal user data
// needed by auth and authz middleware.
type AuthUser struct {
	ID         uuid.UUID      `json:"id"`
	RoleID     *uuid.UUID     `json:"role_id,omitempty"`
	Active     bool           `json:"active"`
	IsApproved bool           `json:"is_approved"`
	Attributes map[string]any `json:"attributes,omitempty"`
}
