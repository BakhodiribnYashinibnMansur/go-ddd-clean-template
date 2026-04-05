package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/admin/featureflag/application"
	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/platform/infrastructure/pgxutil"
)

// ListQuery holds the input for listing feature flags with filtering.
type ListQuery struct {
	Filter domain.FeatureFlagFilter
}

// ListResult holds the output of the list feature flags query.
type ListResult struct {
	Flags []*appdto.FeatureFlagView
	Total int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.FeatureFlagReadRepository
	logger   logger.Log
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.FeatureFlagReadRepository, l logger.Log) *ListHandler {
	return &ListHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListQuery and returns a list of FeatureFlagView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (_ *ListResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListFeatureFlags", "feature_flag")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "List", Entity: "feature_flag", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*appdto.FeatureFlagView, len(views))
	for i, v := range views {
		result[i] = mapToAppView(v)
	}

	return &ListResult{
		Flags: result,
		Total: total,
	}, nil
}
