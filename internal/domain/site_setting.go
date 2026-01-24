package domain

import (
	"time"

	"github.com/google/uuid"
)

// SiteSetting represents a single site configuration setting
type SiteSetting struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"` // string, boolean, integer, json
	Category    string    `json:"category"`   // general, email, maintenance, api
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SiteSettingFilter for querying site settings
type SiteSettingFilter struct {
	ID       *uuid.UUID
	Key      *string
	Category *string
	IsPublic *bool
}

// SiteSettingsFilter for batch queries with pagination
type SiteSettingsFilter struct {
	SiteSettingFilter
	Pagination *Pagination
}

// NewSiteSetting creates a new site setting with defaults
func NewSiteSetting(key, value string) *SiteSetting {
	return &SiteSetting{
		ID:        uuid.New(),
		Key:       key,
		Value:     value,
		ValueType: "string",
		Category:  "general",
		IsPublic:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
