package query

import (
	"context"

	appdto "gct/internal/context/admin/supporting/statistics/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetSecurityStatsQuery is the input for fetching security statistics.
type GetSecurityStatsQuery struct{}

// GetSecurityStatsHandler handles the GetSecurityStatsQuery.
type GetSecurityStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetSecurityStatsHandler creates a new GetSecurityStatsHandler.
func NewGetSecurityStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetSecurityStatsHandler {
	return &GetSecurityStatsHandler{repo: repo, l: l}
}

// Handle executes the GetSecurityStatsQuery and returns a SecurityStatsView.
func (h *GetSecurityStatsHandler) Handle(ctx context.Context, _ GetSecurityStatsQuery) (_ *appdto.SecurityStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSecurityStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetSecurityStats", "statistics")()

	view, err := h.repo.GetSecurityStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetSecurityStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
