package query

import (
	"context"

	appdto "gct/internal/context/admin/statistics/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// GetOverviewQuery is the input for fetching the statistics overview.
type GetOverviewQuery struct{}

// GetOverviewHandler handles the GetOverviewQuery.
type GetOverviewHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetOverviewHandler creates a new GetOverviewHandler.
func NewGetOverviewHandler(repo StatisticsReadRepository, l logger.Log) *GetOverviewHandler {
	return &GetOverviewHandler{repo: repo, l: l}
}

// Handle executes the GetOverviewQuery and returns an OverviewView.
func (h *GetOverviewHandler) Handle(ctx context.Context, _ GetOverviewQuery) (_ *appdto.OverviewView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetOverviewHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetOverview", "statistics")()

	view, err := h.repo.GetOverview(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetOverview failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
