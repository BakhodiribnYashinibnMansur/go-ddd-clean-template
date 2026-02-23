package domain

import (
	"time"

	"github.com/google/uuid"
)

// Translation holds localized data for one entity in one language.
// data JSONB stores field→value pairs e.g. {"title": "Sarlavha", "description": "Tavsif"}
type Translation struct {
	ID         uuid.UUID         `json:"id"`
	EntityType string            `json:"entity_type"`
	EntityID   uuid.UUID         `json:"entity_id"`
	LangCode   string            `json:"lang_code"`
	Data       map[string]string `json:"data"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// EntityTranslations is the response shape: lang_code → field→value
// e.g. {"uz": {"title": "Sarlavha"}, "en": {"title": "Title"}}
type EntityTranslations map[string]map[string]string

// UpsertTranslationsRequest is the PUT request body — same shape as EntityTranslations.
type UpsertTranslationsRequest = EntityTranslations

// TranslationFilter for Gets / Delete.
type TranslationFilter struct {
	EntityType string
	EntityID   uuid.UUID
	LangCode   *string // nil = all languages
}

// Common entity type values (open-ended — any string is valid, these are helpers).
const (
	TranslationEntityRole        = "role"
	TranslationEntityPermission  = "permission"
	TranslationEntityRelation    = "relation"
	TranslationEntityIntegration = "integration"
	TranslationEntitySiteSetting = "site_setting"
	TranslationEntityErrorCode   = "error_code"
)
