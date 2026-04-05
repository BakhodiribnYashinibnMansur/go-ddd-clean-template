package application

import (
	"time"

	"github.com/google/uuid"
)

// UserView is a read-model DTO returned by query handlers.
type UserView struct {
	ID         uuid.UUID      `json:"id"`
	Phone      string         `json:"phone"`
	Email      *string        `json:"email,omitempty"`
	Username   *string        `json:"username,omitempty"`
	RoleID     *uuid.UUID     `json:"role_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Active     bool           `json:"active"`
	IsApproved bool           `json:"is_approved"`
	LastSeen   *time.Time     `json:"last_seen,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// SessionView is a read-model DTO for session data.
type SessionView struct {
	ID           uuid.UUID `json:"id"`
	DeviceType   string    `json:"device_type"`
	DeviceName   string    `json:"device_name"`
	IPAddress    string    `json:"ip_address"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
}
