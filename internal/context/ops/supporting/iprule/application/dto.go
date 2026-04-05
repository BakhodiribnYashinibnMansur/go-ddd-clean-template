package application

import (
	"time"

	"gct/internal/context/ops/supporting/iprule/domain"
)

// IPRuleView is a read-model DTO returned by query handlers.
type IPRuleView struct {
	ID        domain.IPRuleID `json:"id"`
	IPAddress string          `json:"ip_address"`
	Action    string          `json:"action"`
	Reason    string          `json:"reason"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
