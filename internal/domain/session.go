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
	ID               uuid.UUID          `db:"id"                 json:"id"`
	DeviceID         uuid.UUID          `db:"device_id"          json:"device_id"`
	DeviceName       *string            `db:"device_name"        json:"device_name,omitempty"`
	DeviceType       *SessionDeviceType `db:"device_type"        json:"device_type,omitempty"`
	IPAddress        *string            `db:"ip_address"         json:"ip_address,omitempty"`
	UserAgent        *string            `db:"user_agent"         json:"user_agent,omitempty"`
	FCMToken         *string            `db:"fcm_token"          json:"fcm_token,omitempty"`
	UserID           int64              `db:"user_id"            json:"user_id"`
	CompanyID        int64              `db:"company_id"         json:"company_id"`
	RefreshTokenHash string             `db:"refresh_token_hash" json:"-"` // Hashed refresh token
	ExpiresAt        time.Time          `db:"expires_at"         json:"expires_at"`
	LastActivity     time.Time          `db:"last_activity"      json:"last_activity"`
	Revoked          bool               `db:"revoked"            json:"revoked"`
	CreatedAt        time.Time          `db:"created_at"         json:"created_at"`
	UpdatedAt        time.Time          `db:"updated_at"         json:"updated_at"`
}

// SessionFilter represents a filter for session queries. for get and gets endpoints
type SessionFilter struct {
	ID      *uuid.UUID `json:"id,omitempty"`
	UserID  *int64     `json:"user_id,omitempty"`
	Revoked *bool      `json:"revoked,omitempty"`
}

// SessionsFilter represents a filter for multiple sessions with pagination
type SessionsFilter struct {
	SessionFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Helper methods for SessionFilter
func (f SessionFilter) IsIDNull() bool      { return f.ID == nil }
func (f SessionFilter) IsUserIDNull() bool  { return f.UserID == nil }
func (f SessionFilter) IsRevokedNull() bool { return f.Revoked == nil }

// Helper methods for SessionsFilter
func (f SessionsFilter) IsPaginationNull() bool {
	return f.Pagination == nil
}

func (f SessionsFilter) IsValidLimit() bool {
	return !f.IsPaginationNull() && f.Pagination.Limit > 0
}

func (f SessionsFilter) IsValidOffset() bool {
	return !f.IsPaginationNull() && f.Pagination.Offset > 0
}

// Getters for Session
func (s *Session) GetID() uuid.UUID                  { return s.ID }
func (s *Session) GetDeviceID() uuid.UUID            { return s.DeviceID }
func (s *Session) GetDeviceName() *string            { return s.DeviceName }
func (s *Session) GetDeviceType() *SessionDeviceType { return s.DeviceType }
func (s *Session) GetIPAddress() *string             { return s.IPAddress }
func (s *Session) GetUserAgent() *string             { return s.UserAgent }
func (s *Session) GetFCMToken() *string              { return s.FCMToken }
func (s *Session) GetUserID() int64                  { return s.UserID }
func (s *Session) GetCompanyID() int64               { return s.CompanyID }
func (s *Session) GetRefreshTokenHash() string       { return s.RefreshTokenHash }
func (s *Session) GetExpiresAt() time.Time           { return s.ExpiresAt }
func (s *Session) GetLastActivity() time.Time        { return s.LastActivity }
func (s *Session) GetRevoked() bool                  { return s.Revoked }
func (s *Session) GetCreatedAt() time.Time           { return s.CreatedAt }
func (s *Session) GetUpdatedAt() time.Time           { return s.UpdatedAt }

// Setters for Session
func (s *Session) SetDeviceName(deviceName *string) {
	s.DeviceName = deviceName
	s.UpdatedAt = time.Now()
}

func (s *Session) SetDeviceType(deviceType *SessionDeviceType) {
	s.DeviceType = deviceType
	s.UpdatedAt = time.Now()
}
func (s *Session) SetIPAddress(ipAddress *string) { s.IPAddress = ipAddress; s.UpdatedAt = time.Now() }
func (s *Session) SetUserAgent(userAgent *string) { s.UserAgent = userAgent; s.UpdatedAt = time.Now() }
func (s *Session) SetFCMToken(fcmToken *string)   { s.FCMToken = fcmToken; s.UpdatedAt = time.Now() }
func (s *Session) SetRefreshTokenHash(refreshTokenHash string) {
	s.RefreshTokenHash = refreshTokenHash
	s.UpdatedAt = time.Now()
}

func (s *Session) SetExpiresAt(expiresAt time.Time) {
	s.ExpiresAt = expiresAt
	s.UpdatedAt = time.Now()
}
func (s *Session) SetRevoked(revoked bool) { s.Revoked = revoked; s.UpdatedAt = time.Now() }

// Utility methods for Session
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsActive() bool {
	return !s.IsExpired() && !s.Revoked
}

func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
	s.UpdatedAt = time.Now()
}

func (s *Session) Revoke() {
	s.Revoked = true
	s.UpdatedAt = time.Now()
}

func (s *Session) Restore() {
	s.Revoked = false
	s.UpdatedAt = time.Now()
}

func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.UpdatedAt = time.Now()
}

// Getters and Setters for SessionFilter
func (f *SessionFilter) GetID() *uuid.UUID  { return f.ID }
func (f *SessionFilter) SetID(id uuid.UUID) { f.ID = &id }
