package dto

import (
	"time"

	"github.com/google/uuid"
)

// SessionView is a read-model DTO for session data.
// It is a wire-format output DTO: identifiers remain raw uuid.UUID so it can
// be serialized directly to HTTP responses.
type SessionView struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	Revoked      bool      `json:"revoked"`
	CreatedAt    time.Time `json:"created_at"`
}

// SessionsFilter holds optional filters for listing sessions.
type SessionsFilter struct {
	UserID  *uuid.UUID
	Revoked *bool
	Limit   int64
	Offset  int64
}
