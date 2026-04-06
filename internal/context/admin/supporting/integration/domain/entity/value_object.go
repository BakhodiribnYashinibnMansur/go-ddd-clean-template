package entity

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationFilter carries optional filtering parameters for listing integrations.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type IntegrationFilter struct {
	Search  *string
	Type    *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// IntegrationView is a read-model projection optimized for query responses.
// Note: APIKey is included in the view — callers should mask or redact it before exposing to non-admin clients.
type IntegrationView struct {
	ID         IntegrationID     `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	APIKey     string            `json:"api_key"`
	WebhookURL string            `json:"webhook_url"`
	Enabled    bool              `json:"enabled"`
	Config     map[string]string `json:"config"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`

	// JWT half (admin UI surfaces these). The API key hash is intentionally
	// excluded — only its presence matters at the view layer (HasJWT).
	HasJWT                  bool          `json:"has_jwt"`
	JWTAccessTTL            time.Duration `json:"jwt_access_ttl"`
	JWTRefreshTTL           time.Duration `json:"jwt_refresh_ttl"`
	JWTPublicKeyPEM         string        `json:"jwt_public_key_pem"`
	JWTPreviousPublicKeyPEM string        `json:"jwt_previous_public_key_pem"`
	JWTKeyID                string        `json:"jwt_key_id"`
	JWTPreviousKeyID        string        `json:"jwt_previous_key_id"`
	JWTRotatedAt            *time.Time    `json:"jwt_rotated_at,omitempty"`
	JWTRotateEveryDays      int           `json:"jwt_rotate_every_days"`
	JWTBindingMode          string        `json:"jwt_binding_mode"`
	JWTMaxSessions          int           `json:"jwt_max_sessions"`
}

// IntegrationAPIKeyView is a read-model projection for API key validation.
type IntegrationAPIKeyView struct {
	ID            uuid.UUID
	IntegrationID IntegrationID
	Key           string
	Active        bool
}

// JWTIntegrationView is the authoritative lookup result for the ResolveJWTAPIKey
// query. Only includes fields needed on the hot path.
type JWTIntegrationView struct {
	ID                   IntegrationID
	Name                 string
	AccessTTL            time.Duration
	RefreshTTL           time.Duration
	PublicKeyPEM         string
	PreviousPublicKeyPEM string
	KeyID                string
	PreviousKeyID        string
	BindingMode          string
	MaxSessions          int
	RotatedAt            *time.Time
	RotateEveryDays      int
}
