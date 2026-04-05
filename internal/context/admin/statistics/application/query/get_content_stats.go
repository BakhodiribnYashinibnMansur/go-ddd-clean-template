package query

import (
	"context"

	appdto "gct/internal/context/admin/statistics/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// GetContentStatsQuery is the input for fetching content statistics.
type GetContentStatsQuery struct{}

// GetContentStatsHandler handles the GetContentStatsQuery.
type GetContentStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetContentStatsHandler creates a new GetContentStatsHandler.
func NewGetContentStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetContentStatsHandler {
	return &GetContentStatsHandler{repo: repo, l: l}
}

// Handle executes the GetContentStatsQuery and returns a ContentStatsView.
func (h *GetContentStatsHandler) Handle(ctx context.Context, _ GetContentStatsQuery) (_ *appdto.ContentStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetContentStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetContentStats", "statistics")()

	view, err := h.repo.GetContentStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetContentStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
