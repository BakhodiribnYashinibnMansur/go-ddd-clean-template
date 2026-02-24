package iprule

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateIPRuleRequest) (*domain.IPRule, error) {
	ip := &domain.IPRule{
		ID:        uuid.New(),
		IPAddress: req.IPAddress,
		Type:      req.Type,
		Reason:    req.Reason,
		IsActive:  req.IsActive,
	}
	if err := uc.repo.Create(ctx, ip); err != nil {
		return nil, err
	}
	return ip, nil
}
