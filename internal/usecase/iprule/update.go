package iprule

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateIPRuleRequest) (*domain.IPRule, error) {
	ip, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.IPAddress != nil {
		ip.IPAddress = *req.IPAddress
	}
	if req.Type != nil {
		ip.Type = *req.Type
	}
	if req.Reason != nil {
		ip.Reason = *req.Reason
	}
	if req.IsActive != nil {
		ip.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, ip); err != nil {
		return nil, err
	}
	return ip, nil
}
