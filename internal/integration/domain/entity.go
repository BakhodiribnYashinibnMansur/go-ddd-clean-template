package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Integration is the aggregate root for third-party integration management.
type Integration struct {
	shared.AggregateRoot
	name       string
	intType    string
	apiKey     string
	webhookURL string
	enabled    bool
	config     map[string]any
}

// NewIntegration creates a new Integration aggregate.
func NewIntegration(name, intType, apiKey, webhookURL string, enabled bool, config map[string]any) *Integration {
	if config == nil {
		config = make(map[string]any)
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
	return i
}

// ReconstructIntegration rebuilds an Integration aggregate from persisted data. No events are raised.
func ReconstructIntegration(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, intType, apiKey, webhookURL string,
	enabled bool,
	config map[string]any,
) *Integration {
	if config == nil {
		config = make(map[string]any)
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

// UpdateDetails updates mutable fields.
func (i *Integration) UpdateDetails(name, intType, apiKey, webhookURL *string, enabled *bool, config *map[string]any) {
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
func (i *Integration) Config() map[string]any { return i.config }
