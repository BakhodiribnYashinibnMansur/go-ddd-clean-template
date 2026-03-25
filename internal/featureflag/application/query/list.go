package query

import (
	"context"

	appdto "gct/internal/featureflag/application"
	"gct/internal/featureflag/domain"
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
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.FeatureFlagReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of FeatureFlagView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.FeatureFlagView, len(views))
	for i, v := range views {
		result[i] = &appdto.FeatureFlagView{
			ID:                v.ID,
			Name:              v.Name,
			Description:       v.Description,
			Enabled:           v.Enabled,
			RolloutPercentage: v.RolloutPercentage,
		}
	}

	return &ListResult{
		Flags: result,
		Total: total,
	}, nil
}
