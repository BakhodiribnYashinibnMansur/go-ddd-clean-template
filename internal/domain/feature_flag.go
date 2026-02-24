package domain

import (
	"time"

	"github.com/google/uuid"
)

type FeatureFlag struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Type        string     `json:"type"` // bool, string, int, json
	Value       string     `json:"value"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type FeatureFlagFilter struct {
	Search   string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateFeatureFlagRequest struct {
	Key         string `json:"key" binding:"required,min=2,max=100"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Type        string `json:"type" binding:"required,oneof=bool string int json"`
	Value       string `json:"value"`
	Description string `json:"description" binding:"max=500"`
	IsActive    bool   `json:"is_active"`
}

type UpdateFeatureFlagRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Type        *string `json:"type" binding:"omitempty,oneof=bool string int json"`
	Value       *string `json:"value"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active"`
}
