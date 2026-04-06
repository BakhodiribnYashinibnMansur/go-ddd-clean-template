package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/supporting/iprule/application/dto"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListIPRulesQuery holds the input for listing IP rules.
type ListIPRulesQuery struct {
	Filter iprulerepo.IPRuleFilter
}

// ListIPRulesResult holds the output of the list IP rules query.
type ListIPRulesResult struct {
	IPRules []*dto.IPRuleView
	Total   int64
}

// ListIPRulesHandler handles the ListIPRulesQuery.
type ListIPRulesHandler struct {
	readRepo iprulerepo.IPRuleReadRepository
	logger   logger.Log
}

// NewListIPRulesHandler creates a new ListIPRulesHandler.
func NewListIPRulesHandler(readRepo iprulerepo.IPRuleReadRepository, l logger.Log) *ListIPRulesHandler {
	return &ListIPRulesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListIPRulesQuery and returns a list of IPRuleView with total count.
func (h *ListIPRulesHandler) Handle(ctx context.Context, q ListIPRulesQuery) (result *ListIPRulesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListIPRulesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListIPRules", "ip_rule")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListIPRules", Entity: "ip_rule", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*dto.IPRuleView, len(views))
	for i, v := range views {
		items[i] = &dto.IPRuleView{
			ID:        uuid.UUID(v.ID),
			IPAddress: v.IPAddress,
			Action:    v.Action,
			Reason:    v.Reason,
			ExpiresAt: v.ExpiresAt,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListIPRulesResult{
		IPRules: items,
		Total:   total,
	}, nil
}
