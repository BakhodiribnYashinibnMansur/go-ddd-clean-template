package domain

import (
	"encoding/json"
	"time"

	"gct/pkg/validation"

	"github.com/google/uuid"
)

// RawMessage represents a raw JSON message for Swagger documentation
// swagger:type object
// swagger:model RawMessage
type RawMessage = json.RawMessage

// SessionDeviceType represents the type of device.
type SessionDeviceType string

const (
	DeviceTypeDesktop SessionDeviceType = "DESKTOP"
	DeviceTypeMobile  SessionDeviceType = "MOBILE"
	DeviceTypeTablet  SessionDeviceType = "TABLET"
	DeviceTypeBot     SessionDeviceType = "BOT"
	DeviceTypeTV      SessionDeviceType = "TV"
)

func (s SessionDeviceType) IsValid() bool {
	return validation.IsEnumValid(s, []SessionDeviceType{
		DeviceTypeDesktop,
		DeviceTypeMobile,
		DeviceTypeTablet,
		DeviceTypeBot,
		DeviceTypeTV,
	})
}

// Session represents a user session.
type Session struct {
	ID               uuid.UUID          `db:"id"                 json:"id"                        example:"770e8400-e29b-41d4-a716-446655440000"`
	UserID           uuid.UUID          `db:"user_id"            json:"user_id"                   example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceID         uuid.UUID          `db:"device_id"          json:"device_id"                 example:"880e8400-e29b-41d4-a716-446655440000"`
	DeviceName       *string            `db:"device_name"        json:"device_name,omitempty"     example:"iPhone 14 Pro"`
	DeviceType       *SessionDeviceType `db:"device_type"        json:"device_type,omitempty"     example:"MOBILE"  enums:"DESKTOP,MOBILE,TABLET,BOT,TV"`
	IPAddress        *string            `db:"ip_address"         json:"ip_address,omitempty"      example:"192.168.1.1"`
	UserAgent        *string            `db:"user_agent"         json:"user_agent,omitempty"      example:"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)"`
	OS               *string            `db:"os"                 json:"os,omitempty"              example:"iOS"`
	OSVersion        *string            `db:"os_version"         json:"os_version,omitempty"      example:"16.0"`
	Browser          *string            `db:"browser"            json:"browser,omitempty"         example:"Safari"`
	BrowserVersion   *string            `db:"browser_version"    json:"browser_version,omitempty" example:"16.0"`
	FCMToken         *string            `db:"fcm_token"          json:"fcm_token,omitempty"       example:"fcm_token_example_123"`
	RefreshTokenHash string             `db:"refresh_token_hash" json:"refresh_token_hash"        example:"$2a$10$N9qo8uLOickgx2ZMRZoMye"`
	ExpiresAt        time.Time          `db:"expires_at"         json:"expires_at"                example:"2024-02-01T00:00:00Z"  format:"date-time"`
	LastActivity     time.Time          `db:"last_activity"      json:"last_activity"             example:"2024-01-25T10:30:00Z"  format:"date-time"`
	Revoked          bool               `db:"revoked"            json:"revoked"                   example:"false"`
	CreatedAt        time.Time          `db:"created_at"         json:"created_at"                example:"2024-01-01T00:00:00Z"  format:"date-time"`
	UpdatedAt        time.Time          `db:"updated_at"         json:"updated_at"                example:"2024-01-25T10:30:00Z"  format:"date-time"`
}

// Session represents a user session.
// SessionIn represents session-related input data
type SessionIn struct {
	DeviceID       uuid.UUID `db:"device_id"          json:"-"`
	DeviceName     string    `db:"device_name"        json:"device_name,omitempty"`
	DeviceType     string    `db:"device_type"        json:"device_type,omitempty"`
	IP             string    `db:"ip_address"         json:"-"`
	UserAgent      string    `db:"user_agent"         json:"-"`
	OS             string    `db:"os"                 json:"os,omitempty"`
	OSVersion      string    `db:"os_version"         json:"os_version,omitempty"`
	Browser        string    `db:"browser"            json:"browser,omitempty"`
	BrowserVersion string    `db:"browser_version"    json:"browser_version,omitempty"`
	FCMToken       string    `db:"fcm_token"          json:"fcm_token,omitempty"`
}

// SessionFilter represents a filter for session queries. for get and gets endpoints
type SessionFilter struct {
	ID      *uuid.UUID `json:"id,omitempty"`
	UserID  *uuid.UUID `json:"user_id,omitempty"`
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
	return !f.IsPaginationNull() && f.Pagination.Limit > 0 && f.Pagination.Limit <= 1000
}

func (f SessionsFilter) IsValidOffset() bool {
	return !f.IsPaginationNull() && f.Pagination.Offset > 0
}

// Getters for Session
func (s *Session) GetID() uuid.UUID                  { return s.ID }
func (s *Session) GetUserID() uuid.UUID              { return s.UserID }
func (s *Session) GetDeviceID() uuid.UUID            { return s.DeviceID }
func (s *Session) GetDeviceName() *string            { return s.DeviceName }
func (s *Session) GetDeviceType() *SessionDeviceType { return s.DeviceType }
func (s *Session) GetIPAddress() *string             { return s.IPAddress }
func (s *Session) GetUserAgent() *string             { return s.UserAgent }
func (s *Session) GetFCMToken() *string              { return s.FCMToken }
func (s *Session) GetRefreshTokenHash() string       { return s.RefreshTokenHash }

// func (s *Session) GetData() RawMessage               { return s.Data }
func (s *Session) GetExpiresAt() time.Time    { return s.ExpiresAt }
func (s *Session) GetLastActivity() time.Time { return s.LastActivity }
func (s *Session) GetRevoked() bool           { return s.Revoked }
func (s *Session) GetCreatedAt() time.Time    { return s.CreatedAt }
func (s *Session) GetUpdatedAt() time.Time    { return s.UpdatedAt }

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

// func (s *Session) SetData(data RawMessage) {
// 	s.Data = data
// 	s.UpdatedAt = time.Now()
// }

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
	// Extend session expiry by 7 days on each activity
	s.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
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

// SessionContext represents data stored in session.Data JSONB field
// This should contain only frequently-accessed, small data needed on every request
type SessionContext struct {
	RoleID      *uuid.UUID `json:"role_id,omitempty"`
	Language    string     `json:"language,omitempty"`
	Theme       string     `json:"theme,omitempty"`
	TwoFAPassed bool       `json:"2fa_passed,omitempty"`
}

// GetContext and SetContext removed as Data field is removed
// func (s *Session) GetContext() (*SessionContext, error) { ... }
// func (s *Session) SetContext(ctx *SessionContext) error { ... }
