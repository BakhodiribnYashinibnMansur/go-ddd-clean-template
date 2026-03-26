package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// SiteSetting is the aggregate root for global configuration key-value pairs (e.g., maintenance mode, site name).
// Keys are opaque strings; the settingType field categorizes them for UI grouping (e.g., "general", "email").
// All fields are unexported to enforce invariants through domain methods.
type SiteSetting struct {
	shared.AggregateRoot
	key         string
	value       string
	settingType string
	description string
}

// NewSiteSetting creates a new SiteSetting aggregate.
func NewSiteSetting(key, value, settingType, description string) *SiteSetting {
	return &SiteSetting{
		AggregateRoot: shared.NewAggregateRoot(),
		key:           key,
		value:         value,
		settingType:   settingType,
		description:   description,
	}
}

// ReconstructSiteSetting rebuilds a SiteSetting from persisted data. No events are raised.
func ReconstructSiteSetting(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	key, value, settingType, description string,
) *SiteSetting {
	return &SiteSetting{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		key:           key,
		value:         value,
		settingType:   settingType,
		description:   description,
	}
}

// Update applies partial modifications using pointer semantics — nil fields are left unchanged.
// Raises a SettingUpdated event so downstream caches or projections can refresh.
func (s *SiteSetting) Update(key, value, settingType, description *string) {
	if key != nil {
		s.key = *key
	}
	if value != nil {
		s.value = *value
	}
	if settingType != nil {
		s.settingType = *settingType
	}
	if description != nil {
		s.description = *description
	}
	s.Touch()
	s.AddEvent(NewSettingUpdated(s.ID()))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (s *SiteSetting) Key() string         { return s.key }
func (s *SiteSetting) Value() string       { return s.value }
func (s *SiteSetting) Type() string        { return s.settingType }
func (s *SiteSetting) Description() string { return s.description }
