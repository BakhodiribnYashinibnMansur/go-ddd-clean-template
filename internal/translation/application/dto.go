package application

import (
	"time"

	"github.com/google/uuid"
)

// TranslationView is a read-model DTO returned by query handlers.
type TranslationView struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	Language  string    `json:"language"`
	Value     string    `json:"value"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
