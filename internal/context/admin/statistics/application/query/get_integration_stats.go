package query

import (
	"context"

	appdto "gct/internal/context/admin/statistics/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// GetIntegrationStatsQuery is the input for fetching integration statistics.
type GetIntegrationStatsQuery struct{}

// GetIntegrationStatsHandler handles the GetIntegrationStatsQuery.
type GetIntegrationStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetIntegrationStatsHandler creates a new GetIntegrationStatsHandler.
func NewGetIntegrationStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetIntegrationStatsHandler {
	return &GetIntegrationStatsHandler{repo: repo, l: l}
}

// Handle executes the GetIntegrationStatsQuery and returns an IntegrationStatsView.
func (h *GetIntegrationStatsHandler) Handle(ctx context.Context, _ GetIntegrationStatsQuery) (_ *appdto.IntegrationStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetIntegrationStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetIntegrationStats", "statistics")()

	view, err := h.repo.GetIntegrationStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetIntegrationStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
