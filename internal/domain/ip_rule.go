package domain

import (
	"time"

	"github.com/google/uuid"
)

type IPRule struct {
	ID        uuid.UUID `json:"id"`
	IPAddress string    `json:"ip_address"`
	Type      string    `json:"type"` // allow, block
	Reason    string    `json:"reason"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IPRuleFilter struct {
	Search   string
	Type     string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateIPRuleRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=allow block"`
	Reason    string `json:"reason" binding:"max=500"`
	IsActive  bool   `json:"is_active"`
}

type UpdateIPRuleRequest struct {
	IPAddress *string `json:"ip_address"`
	Type      *string `json:"type" binding:"omitempty,oneof=allow block"`
	Reason    *string `json:"reason" binding:"omitempty,max=500"`
	IsActive  *bool   `json:"is_active"`
}
