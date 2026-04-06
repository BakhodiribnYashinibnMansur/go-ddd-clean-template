package domain

import (
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// SessionDeviceType
// ---------------------------------------------------------------------------

// SessionDeviceType classifies the client device. Values must be UPPERCASE to match the PostgreSQL ENUM.
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

// Session is a child entity owned by the User aggregate. It must never be persisted or queried
// independently — all session mutations flow through the User aggregate root to maintain invariants.
// Sessions have a sliding 7-day expiry window that resets on each activity update.
type Session struct {
	shared.BaseEntity
	userID           uuid.UUID
	deviceID         string
	deviceName       string
	deviceType       SessionDeviceType
	ipAddress        shared.IPAddress
	userAgent        shared.UserAgent
	refreshTokenHash    string
	previousRefreshHash string
	deviceFingerprint   string
	expiresAt           time.Time
	lastActivity     time.Time
	revoked          bool
	integrationName  string
}

const defaultSessionDuration = 7 * 24 * time.Hour // 7 days

// DefaultIntegrationName is used when no specific integration is supplied.
// It binds the session to the canonical first-party client.
const DefaultIntegrationName = "gct-client"

// NewSession creates a new session for the given user.
// It validates the IP address via shared.NewIPAddress and normalises the user agent;
// an invalid IP returns shared.ErrInvalidIPAddress and no session is created.
// If integrationName is empty, DefaultIntegrationName is used.
func NewSession(userID uuid.UUID, deviceType SessionDeviceType, ip, userAgent, integrationName string, deviceFingerprint ...string) (*Session, error) {
	ipVO, err := shared.NewIPAddress(ip)
	if err != nil {
		return nil, err
	}
	uaVO := shared.NewUserAgent(userAgent)
	if integrationName == "" {
		integrationName = DefaultIntegrationName
	}
	var fp string
	if len(deviceFingerprint) > 0 {
		fp = deviceFingerprint[0]
	}
	now := time.Now()
	return &Session{
		BaseEntity:        shared.NewBaseEntity(),
		userID:            userID,
		deviceID:          uuid.New().String(),
		deviceType:        deviceType,
		ipAddress:         ipVO,
		userAgent:         uaVO,
		deviceFingerprint: fp,
		expiresAt:         now.Add(defaultSessionDuration),
		lastActivity:      now,
		integrationName:   integrationName,
	}, nil
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
	integrationName string,
	opts ...string,
) *Session {
	ipVO, _ := shared.NewIPAddress(ipAddress) // tolerate legacy/empty rows
	uaVO := shared.NewUserAgent(userAgent)
	if integrationName == "" {
		integrationName = DefaultIntegrationName
	}
	var prevHash string
	if len(opts) > 0 {
		prevHash = opts[0]
	}
	var fp string
	if len(opts) > 1 {
		fp = opts[1]
	}
	return &Session{
		BaseEntity:          shared.NewBaseEntityWithID(id, createdAt, updatedAt, deletedAt),
		userID:              userID,
		deviceID:            deviceID,
		deviceName:          deviceName,
		deviceType:          deviceType,
		ipAddress:           ipVO,
		userAgent:           uaVO,
		refreshTokenHash:    refreshTokenHash,
		previousRefreshHash: prevHash,
		deviceFingerprint:   fp,
		expiresAt:           expiresAt,
		lastActivity:        lastActivity,
		revoked:             revoked,
		integrationName:     integrationName,
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
// Called by middleware on each authenticated request to implement sliding session expiry.
func (s *Session) UpdateActivity() {
	s.lastActivity = time.Now()
	s.expiresAt = time.Now().Add(defaultSessionDuration)
	s.Touch()
}

// SetRefreshTokenHash stores the hashed refresh token for rotation verification.
// The raw refresh token is never stored — only its hash, for constant-time comparison during rotation.
func (s *Session) SetRefreshTokenHash(hash string) {
	s.refreshTokenHash = hash
	s.Touch()
}

// RotateRefreshHash sets a new refresh token hash and moves the current
// hash to the previous slot. Returns the old hash (now in previous slot).
// This enables one-generation reuse detection.
func (s *Session) RotateRefreshHash(newHash string) string {
	old := s.refreshTokenHash
	s.previousRefreshHash = old
	s.refreshTokenHash = newHash
	s.Touch()
	return old
}

// SetIntegrationName assigns the integration (audience) this session is bound to.
// It is idempotent and does not bump the modification timestamp on its own.
func (s *Session) SetIntegrationName(name string) {
	if name == "" {
		name = DefaultIntegrationName
	}
	s.integrationName = name
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (s *Session) UserID() uuid.UUID             { return s.userID }
func (s *Session) DeviceID() string              { return s.deviceID }
func (s *Session) DeviceName() string            { return s.deviceName }
func (s *Session) DeviceType() SessionDeviceType { return s.deviceType }
func (s *Session) IPAddress() shared.IPAddress   { return s.ipAddress }
func (s *Session) UserAgent() shared.UserAgent   { return s.userAgent }
func (s *Session) RefreshTokenHash() string      { return s.refreshTokenHash }
func (s *Session) PreviousRefreshHash() string   { return s.previousRefreshHash }
func (s *Session) ExpiresAt() time.Time          { return s.expiresAt }
func (s *Session) LastActivity() time.Time       { return s.lastActivity }
func (s *Session) IsRevoked() bool               { return s.revoked }
func (s *Session) IntegrationName() string       { return s.integrationName }
func (s *Session) DeviceFingerprint() string     { return s.deviceFingerprint }
