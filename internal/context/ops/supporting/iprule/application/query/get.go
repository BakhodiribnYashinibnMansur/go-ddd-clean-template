package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/supporting/iprule/application/dto"
	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetIPRuleQuery holds the input for getting a single IP rule.
type GetIPRuleQuery struct {
	ID ipruleentity.IPRuleID
}

// GetIPRuleHandler handles the GetIPRuleQuery.
type GetIPRuleHandler struct {
	readRepo iprulerepo.IPRuleReadRepository
	logger   logger.Log
}

// NewGetIPRuleHandler creates a new GetIPRuleHandler.
func NewGetIPRuleHandler(readRepo iprulerepo.IPRuleReadRepository, l logger.Log) *GetIPRuleHandler {
	return &GetIPRuleHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetIPRuleQuery and returns an IPRuleView.
func (h *GetIPRuleHandler) Handle(ctx context.Context, q GetIPRuleQuery) (result *dto.IPRuleView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetIPRule", "ip_rule")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetIPRule", Entity: "ip_rule", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.IPRuleView{
		ID:        uuid.UUID(v.ID),
		IPAddress: v.IPAddress,
		Action:    v.Action,
		Reason:    v.Reason,
		ExpiresAt: v.ExpiresAt,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
