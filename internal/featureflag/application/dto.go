package application

import (
	"time"

	"github.com/google/uuid"
)

// FeatureFlagView is a read-model DTO returned by query handlers.
type FeatureFlagView struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Enabled           bool      `json:"enabled"`
	RolloutPercentage int       `json:"rollout_percentage"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
