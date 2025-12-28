package domain

import (
	"time"

	"github.com/google/uuid"
)

// SessionDeviceType represents the type of device.
type SessionDeviceType string

const (
	DeviceTypeDesktop SessionDeviceType = "DESKTOP"
	DeviceTypeMobile  SessionDeviceType = "MOBILE"
	DeviceTypeTablet  SessionDeviceType = "TABLET"
	DeviceTypeBot     SessionDeviceType = "BOT"
	DeviceTypeTV      SessionDeviceType = "TV"
)

// Session represents a user session.
type Session struct {
	ID               uuid.UUID          `json:"id" db:"id"`
	DeviceID         uuid.UUID          `json:"device_id" db:"device_id"`
	DeviceName       *string            `json:"device_name,omitempty" db:"device_name"`
	DeviceType       *SessionDeviceType `json:"device_type,omitempty" db:"device_type"`
	IPAddress        *string            `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent        *string            `json:"user_agent,omitempty" db:"user_agent"`
	FCMToken         *string            `json:"fcm_token,omitempty" db:"fcm_token"`
	UserID           int64              `json:"user_id" db:"user_id"`
	CompanyID        int64              `json:"company_id" db:"company_id"`
	RefreshTokenHash string             `json:"-" db:"refresh_token_hash"` // Hashed refresh token
	ExpiresAt        time.Time          `json:"expires_at" db:"expires_at"`
	LastActivity     time.Time          `json:"last_activity" db:"last_activity"`
	Revoked          bool               `json:"revoked" db:"revoked"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
}

// SessionFilter represents a filter for session queries. for get and gets endpoints
type SessionFilter struct {
	ID uuid.UUID
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is still active.
func (s *Session) IsActive() bool {
	return !s.IsExpired()
}

// UpdateActivity updates the last activity timestamp.
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}
