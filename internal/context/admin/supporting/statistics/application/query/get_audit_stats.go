package query

import (
	"context"

	"gct/internal/context/admin/supporting/statistics/application/dto"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetAuditStatsQuery is the input for fetching audit log statistics.
type GetAuditStatsQuery struct{}

// GetAuditStatsHandler handles the GetAuditStatsQuery.
type GetAuditStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetAuditStatsHandler creates a new GetAuditStatsHandler.
func NewGetAuditStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetAuditStatsHandler {
	return &GetAuditStatsHandler{repo: repo, l: l}
}

// Handle executes the GetAuditStatsQuery and returns an AuditStatsView.
func (h *GetAuditStatsHandler) Handle(ctx context.Context, _ GetAuditStatsQuery) (_ *dto.AuditStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetAuditStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetAuditStats", "statistics")()

	view, err := h.repo.GetAuditStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetAuditStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
