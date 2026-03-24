package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// SessionDeviceType
// ---------------------------------------------------------------------------

type SessionDeviceType string

const (
	DeviceDesktop SessionDeviceType = "DESKTOP"
	DeviceMobile  SessionDeviceType = "MOBILE"
	DeviceTablet  SessionDeviceType = "TABLET"
	DeviceBot     SessionDeviceType = "BOT"
	DeviceTV      SessionDeviceType = "TV"
)

// ---------------------------------------------------------------------------
// Session — child entity within the User aggregate
// ---------------------------------------------------------------------------

type Session struct {
	shared.BaseEntity
	userID           uuid.UUID
	deviceID         string
	deviceName       string
	deviceType       SessionDeviceType
	ipAddress        string
	userAgent        string
	refreshTokenHash string
	expiresAt        time.Time
	lastActivity     time.Time
	revoked          bool
}

const defaultSessionDuration = 7 * 24 * time.Hour // 7 days

// NewSession creates a new session for the given user.
func NewSession(userID uuid.UUID, deviceType SessionDeviceType, ip, userAgent string) *Session {
	now := time.Now()
	return &Session{
		BaseEntity:   shared.NewBaseEntity(),
		userID:       userID,
		deviceID:     uuid.New().String(),
		deviceType:   deviceType,
		ipAddress:    ip,
		userAgent:    userAgent,
		expiresAt:    now.Add(defaultSessionDuration),
		lastActivity: now,
	}
}

// ReconstructSession rebuilds a Session from persisted data.
func ReconstructSession(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	userID uuid.UUID,
	deviceID, deviceName string,
	deviceType SessionDeviceType,
	ipAddress, userAgent, refreshTokenHash string,
	expiresAt, lastActivity time.Time,
	revoked bool,
) *Session {
	return &Session{
		BaseEntity:       shared.NewBaseEntityWithID(id, createdAt, updatedAt, deletedAt),
		userID:           userID,
		deviceID:         deviceID,
		deviceName:       deviceName,
		deviceType:       deviceType,
		ipAddress:        ipAddress,
		userAgent:        userAgent,
		refreshTokenHash: refreshTokenHash,
		expiresAt:        expiresAt,
		lastActivity:     lastActivity,
		revoked:          revoked,
	}
}

// ---------------------------------------------------------------------------
// Behaviour
// ---------------------------------------------------------------------------

// IsExpired returns true if the session has passed its expiry time.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// IsActive returns true if the session is neither expired nor revoked.
func (s *Session) IsActive() bool {
	return !s.IsExpired() && !s.revoked
}

// Revoke marks the session as revoked.
func (s *Session) Revoke() {
	s.revoked = true
	s.Touch()
}

// UpdateActivity refreshes the last activity timestamp and extends expiry by 7 days.
func (s *Session) UpdateActivity() {
	s.lastActivity = time.Now()
	s.expiresAt = time.Now().Add(defaultSessionDuration)
	s.Touch()
}

// SetRefreshTokenHash stores the hashed refresh token.
func (s *Session) SetRefreshTokenHash(hash string) {
	s.refreshTokenHash = hash
	s.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (s *Session) UserID() uuid.UUID            { return s.userID }
func (s *Session) DeviceID() string             { return s.deviceID }
func (s *Session) DeviceName() string           { return s.deviceName }
func (s *Session) DeviceType() SessionDeviceType { return s.deviceType }
func (s *Session) IPAddress() string            { return s.ipAddress }
func (s *Session) UserAgent() string            { return s.userAgent }
func (s *Session) RefreshTokenHash() string     { return s.refreshTokenHash }
func (s *Session) ExpiresAt() time.Time         { return s.expiresAt }
func (s *Session) LastActivity() time.Time      { return s.lastActivity }
func (s *Session) IsRevoked() bool              { return s.revoked }
