package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/ops/iprule/application"
	"gct/internal/context/ops/iprule/domain"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetIPRuleQuery holds the input for getting a single IP rule.
type GetIPRuleQuery struct {
	ID uuid.UUID
}

// GetIPRuleHandler handles the GetIPRuleQuery.
type GetIPRuleHandler struct {
	readRepo domain.IPRuleReadRepository
	logger   logger.Log
}

// NewGetIPRuleHandler creates a new GetIPRuleHandler.
func NewGetIPRuleHandler(readRepo domain.IPRuleReadRepository, l logger.Log) *GetIPRuleHandler {
	return &GetIPRuleHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetIPRuleQuery and returns an IPRuleView.
func (h *GetIPRuleHandler) Handle(ctx context.Context, q GetIPRuleQuery) (result *appdto.IPRuleView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetIPRule", "ip_rule")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetIPRule", Entity: "ip_rule", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.IPRuleView{
		ID:        v.ID,
		IPAddress: v.IPAddress,
		Action:    v.Action,
		Reason:    v.Reason,
		ExpiresAt: v.ExpiresAt,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
