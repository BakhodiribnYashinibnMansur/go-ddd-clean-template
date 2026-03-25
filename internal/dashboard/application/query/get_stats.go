package query

import (
	"context"

	appdto "gct/internal/dashboard/application"
	"gct/internal/shared/infrastructure/logger"
)

// GetStatsQuery is the input for fetching dashboard statistics.
type GetStatsQuery struct{}

// GetStatsHandler handles the GetStatsQuery.
type GetStatsHandler struct {
	repo DashboardReadRepository
	l    logger.Log
}

// NewGetStatsHandler creates a new GetStatsHandler.
func NewGetStatsHandler(repo DashboardReadRepository, l logger.Log) *GetStatsHandler {
	return &GetStatsHandler{repo: repo, l: l}
}

// Handle executes the GetStatsQuery and returns a DashboardStatsView.
func (h *GetStatsHandler) Handle(ctx context.Context, _ GetStatsQuery) (*appdto.DashboardStatsView, error) {
	view, err := h.repo.GetStats(ctx)
	if err != nil {
		h.l.Errorc(ctx, "dashboard.query.GetStats failed", "error", err)
		return nil, err
	}
	return view, nil
}
