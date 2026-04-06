package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/generic/featureflag/application/dto"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListQuery holds the input for listing feature flags with filtering.
type ListQuery struct {
	Filter ffrepo.FeatureFlagFilter
}

// ListResult holds the output of the list feature flags query.
type ListResult struct {
	Flags []*dto.FeatureFlagView
	Total int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo ffrepo.FeatureFlagReadRepository
	logger   logger.Log
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo ffrepo.FeatureFlagReadRepository, l logger.Log) *ListHandler {
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

	result := make([]*dto.FeatureFlagView, len(views))
	for i, v := range views {
		result[i] = mapToAppView(v)
	}

	return &ListResult{
		Flags: result,
		Total: total,
	}, nil
}
