package query

import (
	"context"

	appdto "gct/internal/context/admin/supporting/statistics/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetUserStatsQuery is the input for fetching user statistics.
type GetUserStatsQuery struct{}

// GetUserStatsHandler handles the GetUserStatsQuery.
type GetUserStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetUserStatsHandler creates a new GetUserStatsHandler.
func NewGetUserStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetUserStatsHandler {
	return &GetUserStatsHandler{repo: repo, l: l}
}

// Handle executes the GetUserStatsQuery and returns a UserStatsView.
func (h *GetUserStatsHandler) Handle(ctx context.Context, _ GetUserStatsQuery) (_ *appdto.UserStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetUserStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetUserStats", "statistics")()

	view, err := h.repo.GetUserStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetUserStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
