package query

import (
	"context"

	appdto "gct/internal/context/admin/statistics/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
)

// GetFeatureFlagStatsQuery is the input for fetching feature flag statistics.
type GetFeatureFlagStatsQuery struct{}

// GetFeatureFlagStatsHandler handles the GetFeatureFlagStatsQuery.
type GetFeatureFlagStatsHandler struct {
	repo StatisticsReadRepository
	l    logger.Log
}

// NewGetFeatureFlagStatsHandler creates a new GetFeatureFlagStatsHandler.
func NewGetFeatureFlagStatsHandler(repo StatisticsReadRepository, l logger.Log) *GetFeatureFlagStatsHandler {
	return &GetFeatureFlagStatsHandler{repo: repo, l: l}
}

// Handle executes the GetFeatureFlagStatsQuery and returns a FeatureFlagStatsView.
func (h *GetFeatureFlagStatsHandler) Handle(ctx context.Context, _ GetFeatureFlagStatsQuery) (_ *appdto.FeatureFlagStatsView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetFeatureFlagStatsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.l, ctx, "GetFeatureFlagStats", "statistics")()

	view, err := h.repo.GetFeatureFlagStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "statistics.query.GetFeatureFlagStats failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
