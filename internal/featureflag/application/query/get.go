package query

import (
	"context"

	appdto "gct/internal/featureflag/application"
	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single feature flag.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.FeatureFlagReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.FeatureFlagReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns a FeatureFlagView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (*appdto.FeatureFlagView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.FeatureFlagView{
		ID:                view.ID,
		Name:              view.Name,
		Description:       view.Description,
		Enabled:           view.Enabled,
		RolloutPercentage: view.RolloutPercentage,
	}, nil
}
