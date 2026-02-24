package iprule

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, ip *domain.IPRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IPRule, error)
	List(ctx context.Context, filter domain.IPRuleFilter) ([]domain.IPRule, int64, error)
	Update(ctx context.Context, ip *domain.IPRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateIPRuleRequest) (*domain.IPRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IPRule, error)
	List(ctx context.Context, filter domain.IPRuleFilter) ([]domain.IPRule, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateIPRuleRequest) (*domain.IPRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
