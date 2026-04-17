package repository

import (
	"context"

	"gct/internal/context/ops/supporting/iprule/domain/entity"
	shareddomain "gct/internal/kernel/domain"
)

// IPRuleFilter carries optional filtering parameters for listing IP rules.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type IPRuleFilter struct {
	IPAddress *string
	Action    *string
	Limit     int64
	Offset    int64
}

// IPRuleRepository is the write-side repository for the IPRule aggregate.
// List is included on the write side because enforcement middleware needs access to full aggregates
// for real-time IP matching, not just read-model projections.
type IPRuleRepository interface {
	Save(ctx context.Context, q shareddomain.Querier, entity *entity.IPRule) error
	FindByID(ctx context.Context, id entity.IPRuleID) (*entity.IPRule, error)
	Update(ctx context.Context, q shareddomain.Querier, entity *entity.IPRule) error
	Delete(ctx context.Context, q shareddomain.Querier, id entity.IPRuleID) error
	List(ctx context.Context, filter IPRuleFilter) ([]*entity.IPRule, int64, error)
}
