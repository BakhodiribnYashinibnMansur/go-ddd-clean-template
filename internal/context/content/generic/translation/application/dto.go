package application

import (
	"time"

	"gct/internal/context/content/generic/translation/domain"
)

// TranslationView is a read-model DTO returned by query handlers.
type TranslationView struct {
	ID        domain.TranslationID `json:"id"`
	Key       string               `json:"key"`
	Language  string               `json:"language"`
	Value     string               `json:"value"`
	Group     string               `json:"group"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}
