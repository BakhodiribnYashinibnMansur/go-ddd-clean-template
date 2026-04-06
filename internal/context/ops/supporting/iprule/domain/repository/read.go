package repository

import (
	"context"
	"time"

	"gct/internal/context/ops/supporting/iprule/domain/entity"
)

// IPRuleView is a read-model projection optimized for query responses and admin UI display.
type IPRuleView struct {
	ID        entity.IPRuleID `json:"id"`
	IPAddress string          `json:"ip_address"`
	Action    string          `json:"action"`
	Reason    string          `json:"reason"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// IPRuleReadRepository is the read-side repository returning projected views.
// Implementations must return ErrIPRuleNotFound when FindByID yields no result.
type IPRuleReadRepository interface {
	FindByID(ctx context.Context, id entity.IPRuleID) (*IPRuleView, error)
	List(ctx context.Context, filter IPRuleFilter) ([]*IPRuleView, int64, error)
}
