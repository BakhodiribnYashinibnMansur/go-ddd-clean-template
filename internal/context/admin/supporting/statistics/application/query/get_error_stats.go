package query

import (
	"context"

	"gct/internal/context/admin/supporting/statistics/application/dto"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetErrorStatsQuery is the input for fetching error statistics.
type GetErrorStatsQuery struct{}

// GetErrorStatsHandler handles the GetErrorStatsQuery.
type GetErrorStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetErrorStatsHandler creates a new GetErrorStatsHandler.
func NewGetErrorStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetErrorStatsHandler {
	return &GetErrorStatsHandler{repo: repo, l: l}
}

// Handle executes the GetErrorStatsQuery and returns an ErrorStatsView.
func (h *GetErrorStatsHandler) Handle(ctx context.Context, _ GetErrorStatsQuery) (_ *dto.ErrorStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetErrorStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetErrorStats", "statistics")()

	view, err := h.repo.GetErrorStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetErrorStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
