package query

import (
	"context"

	appdto "gct/internal/context/admin/supporting/statistics/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetSessionStatsQuery is the input for fetching session statistics.
type GetSessionStatsQuery struct{}

// GetSessionStatsHandler handles the GetSessionStatsQuery.
type GetSessionStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetSessionStatsHandler creates a new GetSessionStatsHandler.
func NewGetSessionStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetSessionStatsHandler {
	return &GetSessionStatsHandler{repo: repo, l: l}
}

// Handle executes the GetSessionStatsQuery and returns a SessionStatsView.
func (h *GetSessionStatsHandler) Handle(ctx context.Context, _ GetSessionStatsQuery) (_ *appdto.SessionStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSessionStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetSessionStats", "statistics")()

	view, err := h.repo.GetSessionStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetSessionStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
