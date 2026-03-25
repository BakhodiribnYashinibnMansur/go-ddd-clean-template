package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IPRuleFilter carries filtering parameters for listing IP rules.
type IPRuleFilter struct {
	IPAddress *string
	Action    *string
	Limit     int64
	Offset    int64
}

// IPRuleRepository is the write-side repository for the IPRule aggregate.
type IPRuleRepository interface {
	Save(ctx context.Context, entity *IPRule) error
	FindByID(ctx context.Context, id uuid.UUID) (*IPRule, error)
	Update(ctx context.Context, entity *IPRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter IPRuleFilter) ([]*IPRule, int64, error)
}

// IPRuleView is a read-model DTO for IP rules.
type IPRuleView struct {
	ID        uuid.UUID  `json:"id"`
	IPAddress string     `json:"ip_address"`
	Action    string     `json:"action"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// IPRuleReadRepository is the read-side repository returning projected views.
type IPRuleReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*IPRuleView, error)
	List(ctx context.Context, filter IPRuleFilter) ([]*IPRuleView, int64, error)
}
