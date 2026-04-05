package domain

import (
	"fmt"
	"strings"
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
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
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		intType:       intType,
		apiKey:        apiKey,
		webhookURL:    webhookURL,
		enabled:       enabled,
		config:        config,
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
) *Integration {
	if config == nil {
		config = make(map[string]string)
	}
	return &Integration{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:          name,
		intType:       intType,
		apiKey:        apiKey,
		webhookURL:    webhookURL,
		enabled:       enabled,
		config:        config,
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

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (i *Integration) Name() string           { return i.name }
func (i *Integration) Type() string           { return i.intType }
func (i *Integration) APIKey() string         { return i.apiKey }
func (i *Integration) WebhookURL() string     { return i.webhookURL }
func (i *Integration) Enabled() bool          { return i.enabled }
func (i *Integration) Config() map[string]string { return i.config }
