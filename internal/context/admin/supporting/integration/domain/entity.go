package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// JWT binding mode constants for Integration.jwtBindingMode.
const (
	BindingModeOff    = "off"
	BindingModeWarn   = "warn"
	BindingModeStrict = "strict"
)

// Integration is the aggregate root for third-party integration management.
// It encapsulates credentials (apiKey) and routing (webhookURL) for external services.
// The config map provides extensibility for integration-specific settings without schema changes.
type Integration struct {
	shared.AggregateRoot
	name       string
	intType    string
	apiKey     string
	webhookURL string
	enabled    bool
	config     map[string]string

	// JWT half — provisioned separately via SetJWT.
	jwtAPIKeyHash           []byte
	jwtAccessTTL            time.Duration
	jwtRefreshTTL           time.Duration
	jwtPublicKeyPEM         string
	jwtPreviousPublicKeyPEM string
	jwtKeyID                string
	jwtPreviousKeyID        string
	jwtRotatedAt            *time.Time
	jwtRotateEveryDays      int
	jwtBindingMode          string
	jwtMaxSessions          int
}

// NewIntegration creates a new Integration aggregate.
// Returns an error if name or intType is empty after trim.
func NewIntegration(name, intType, apiKey, webhookURL string, enabled bool, config map[string]string) (*Integration, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("new_integration: %s", "name is required")
	}
	if strings.TrimSpace(intType) == "" {
		return nil, fmt.Errorf("new_integration: %s", "type is required")
	}
	if config == nil {
		config = make(map[string]string)
	}
	i := &Integration{
		AggregateRoot:      shared.NewAggregateRoot(),
		name:               name,
		intType:            intType,
		apiKey:             apiKey,
		webhookURL:         webhookURL,
		enabled:            enabled,
		config:             config,
		jwtRotateEveryDays: 30,
		jwtBindingMode:     BindingModeWarn,
	}
	i.AddEvent(NewIntegrationConnected(i.ID(), name, intType))
	return i, nil
}

// ReconstructIntegration rebuilds an Integration aggregate from persisted data. No events are raised.
func ReconstructIntegration(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, intType, apiKey, webhookURL string,
	enabled bool,
	config map[string]string,
	jwtAPIKeyHash []byte,
	jwtAccessTTL, jwtRefreshTTL time.Duration,
	jwtPublicKeyPEM, jwtPreviousPublicKeyPEM string,
	jwtKeyID, jwtPreviousKeyID string,
	jwtRotatedAt *time.Time,
	jwtRotateEveryDays int,
	jwtBindingMode string,
	jwtMaxSessions int,
) *Integration {
	if config == nil {
		config = make(map[string]string)
	}
	if jwtBindingMode == "" {
		jwtBindingMode = BindingModeWarn
	}
	return &Integration{
		AggregateRoot:           shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:                    name,
		intType:                 intType,
		apiKey:                  apiKey,
		webhookURL:              webhookURL,
		enabled:                 enabled,
		config:                  config,
		jwtAPIKeyHash:           jwtAPIKeyHash,
		jwtAccessTTL:            jwtAccessTTL,
		jwtRefreshTTL:           jwtRefreshTTL,
		jwtPublicKeyPEM:         jwtPublicKeyPEM,
		jwtPreviousPublicKeyPEM: jwtPreviousPublicKeyPEM,
		jwtKeyID:                jwtKeyID,
		jwtPreviousKeyID:        jwtPreviousKeyID,
		jwtRotatedAt:            jwtRotatedAt,
		jwtRotateEveryDays:      jwtRotateEveryDays,
		jwtBindingMode:          jwtBindingMode,
		jwtMaxSessions:          jwtMaxSessions,
	}
}

// UpdateDetails applies a partial update using pointer-based optionality.
// Nil pointers are skipped, allowing callers to update only the fields they provide.
// Touch is called to advance the updatedAt timestamp for optimistic concurrency.
func (i *Integration) UpdateDetails(name, intType, apiKey, webhookURL *string, enabled *bool, config *map[string]string) {
	if name != nil {
		i.name = *name
	}
	if intType != nil {
		i.intType = *intType
	}
	if apiKey != nil {
		i.apiKey = *apiKey
	}
	if webhookURL != nil {
		i.webhookURL = *webhookURL
	}
	if enabled != nil {
		i.enabled = *enabled
	}
	if config != nil {
		i.config = *config
	}
	i.Touch()
}

// SetJWT configures the JWT half of the integration. Returns an error if
// mode is invalid or TTLs non-positive. Called by the JWT rotation/
// provisioning flow, not the basic CRUD.
func (i *Integration) SetJWT(hash []byte, accessTTL, refreshTTL time.Duration, publicKeyPEM, keyID, mode string, maxSessions int) error {
	if accessTTL <= 0 {
		return errors.New("set_jwt: accessTTL must be positive")
	}
	if refreshTTL <= 0 {
		return errors.New("set_jwt: refreshTTL must be positive")
	}
	switch mode {
	case BindingModeOff, BindingModeWarn, BindingModeStrict:
	default:
		return fmt.Errorf("set_jwt: invalid binding mode %q", mode)
	}
	i.jwtAPIKeyHash = hash
	i.jwtAccessTTL = accessTTL
	i.jwtRefreshTTL = refreshTTL
	i.jwtPublicKeyPEM = publicKeyPEM
	i.jwtKeyID = keyID
	i.jwtBindingMode = mode
	i.jwtMaxSessions = maxSessions
	i.Touch()
	return nil
}

// RotateJWTKey moves the current key material to previous slots and installs
// the new values. Called by the keyring scheduler.
func (i *Integration) RotateJWTKey(newPublicKeyPEM, newKeyID string) {
	i.jwtPreviousPublicKeyPEM = i.jwtPublicKeyPEM
	i.jwtPreviousKeyID = i.jwtKeyID
	i.jwtPublicKeyPEM = newPublicKeyPEM
	i.jwtKeyID = newKeyID
	now := time.Now().UTC()
	i.jwtRotatedAt = &now
	i.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (i *Integration) TypedID() IntegrationID    { return IntegrationID(i.ID()) }
func (i *Integration) Name() string              { return i.name }
func (i *Integration) Type() string              { return i.intType }
func (i *Integration) APIKey() string            { return i.apiKey }
func (i *Integration) WebhookURL() string        { return i.webhookURL }
func (i *Integration) Enabled() bool             { return i.enabled }
func (i *Integration) Config() map[string]string { return i.config }

func (i *Integration) JWTAPIKeyHash() []byte           { return i.jwtAPIKeyHash }
func (i *Integration) JWTAccessTTL() time.Duration     { return i.jwtAccessTTL }
func (i *Integration) JWTRefreshTTL() time.Duration    { return i.jwtRefreshTTL }
func (i *Integration) JWTPublicKeyPEM() string         { return i.jwtPublicKeyPEM }
func (i *Integration) JWTPreviousPublicKeyPEM() string { return i.jwtPreviousPublicKeyPEM }
func (i *Integration) JWTKeyID() string                { return i.jwtKeyID }
func (i *Integration) JWTPreviousKeyID() string        { return i.jwtPreviousKeyID }
func (i *Integration) JWTRotatedAt() *time.Time        { return i.jwtRotatedAt }
func (i *Integration) JWTRotateEveryDays() int         { return i.jwtRotateEveryDays }
func (i *Integration) JWTBindingMode() string          { return i.jwtBindingMode }
func (i *Integration) JWTMaxSessions() int             { return i.jwtMaxSessions }
