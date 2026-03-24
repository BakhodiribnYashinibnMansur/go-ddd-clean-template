package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"gct/internal/shared/domain/consts"

	"github.com/google/uuid"
)

// Integration represents a third-party integration platform configuration.
type Integration struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`               // Integration name (e.g., "Stripe", "PayPal")
	Description string         `json:"description" db:"description"` // Integration description
	BaseURL     string         `json:"base_url" db:"base_url"`       // Base API URL
	IsActive    bool           `json:"is_active" db:"is_active"`     // Whether integration is active
	Config      map[string]any `json:"config" db:"config"`           // Additional configuration (JSON)
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// APIKey represents an API key for accessing integrations.
type APIKey struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	IntegrationID uuid.UUID  `json:"integration_id" db:"integration_id"`   // Foreign key to Integration
	Name          string     `json:"name" db:"name"`                       // Key name/label
	Key           string     `json:"key" db:"key"`                         // The actual API key (hashed)
	KeyPrefix     string     `json:"key_prefix" db:"key_prefix"`           // Visible prefix for identification
	IsActive      bool       `json:"is_active" db:"is_active"`             // Whether key is active
	ExpiresAt     *time.Time `json:"expires_at,omitempty" db:"expires_at"` // Optional expiration
	LastUsedAt    *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// GenerateKey generates a new random API key with a prefix.
func (a *APIKey) GenerateKey() (string, error) {
	b := make([]byte, consts.APIKeyLength)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(consts.APIKeyCharset))))
		if err != nil {
			return "", err
		}
		b[i] = consts.APIKeyCharset[num.Int64()]
	}

	rawKey := fmt.Sprintf("%s_%s", a.KeyPrefix, string(b))
	a.Key = a.HashKey(rawKey)
	return rawKey, nil
}

// HashKey hashes a raw API key using SHA-256.
func (a *APIKey) HashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// ValidateKey compares a raw key with the stored hash.
func (a *APIKey) ValidateKey(rawKey string) bool {
	return a.Key == a.HashKey(rawKey)
}

// IntegrationWithKeys combines Integration with its API keys.
type IntegrationWithKeys struct {
	Integration
	APIKeys []APIKey `json:"api_keys"`
}

// IntegrationFilter defines filtering options for listing integrations.
type IntegrationFilter struct {
	Search   string
	IsActive *bool
	Limit    int
	Offset   int
}

// CreateIntegrationRequest represents the request to create an integration.
type CreateIntegrationRequest struct {
	Name        string         `json:"name" binding:"required,min=3,max=100"`
	Description string         `json:"description" binding:"max=500"`
	BaseURL     string         `json:"base_url" binding:"required,url"`
	IsActive    bool           `json:"is_active"`
	Config      map[string]any `json:"config"`
}

// UpdateIntegrationRequest represents the request to update an integration.
type UpdateIntegrationRequest struct {
	Name        *string         `json:"name" binding:"omitempty,min=3,max=100"`
	Description *string         `json:"description" binding:"omitempty,max=500"`
	BaseURL     *string         `json:"base_url" binding:"omitempty,url"`
	IsActive    *bool           `json:"is_active"`
	Config      *map[string]any `json:"config"`
}

// CreateAPIKeyRequest represents the request to create an API key.
type CreateAPIKeyRequest struct {
	IntegrationID uuid.UUID  `json:"integration_id" binding:"required"`
	Name          string     `json:"name" binding:"required,min=3,max=100"`
	ExpiresAt     *time.Time `json:"expires_at"`
}
