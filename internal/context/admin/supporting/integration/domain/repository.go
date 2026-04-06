package domain

import (
	"context"
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

// IntegrationRepository is the write-side repository for the Integration aggregate.
// Delete performs a hard delete — callers should ensure authorization before invoking.
type IntegrationRepository interface {
	Save(ctx context.Context, entity *Integration) error
	FindByID(ctx context.Context, id IntegrationID) (*Integration, error)
	Update(ctx context.Context, entity *Integration) error
	Delete(ctx context.Context, id IntegrationID) error
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

// IntegrationReadRepository is the read-side repository returning projected views.
// Implementations must return ErrIntegrationNotFound when FindByID yields no result.
type IntegrationReadRepository interface {
	FindByID(ctx context.Context, id IntegrationID) (*IntegrationView, error)
	List(ctx context.Context, filter IntegrationFilter) ([]*IntegrationView, int64, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*IntegrationAPIKeyView, error)

	// ListActiveJWT returns all integrations that have jwt_api_key_hash set
	// (NULL means JWT is not provisioned for that integration yet).
	ListActiveJWT(ctx context.Context) ([]JWTIntegrationView, error)

	// FindJWTByHash returns the integration whose jwt_api_key_hash exactly
	// matches the provided hash. Uses the DB unique index. Returns
	// ErrIntegrationNotFound if not found.
	FindJWTByHash(ctx context.Context, hash []byte) (*JWTIntegrationView, error)
}
