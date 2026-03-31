package query

import (
	"context"
	"time"

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

	return mapToAppView(view), nil
}

// mapToAppView converts a domain FeatureFlagView to an application FeatureFlagView.
func mapToAppView(v *domain.FeatureFlagView) *appdto.FeatureFlagView {
	ruleGroups := make([]appdto.RuleGroupView, len(v.RuleGroups))
	for i, rg := range v.RuleGroups {
		conditions := make([]appdto.ConditionView, len(rg.Conditions))
		for j, c := range rg.Conditions {
			conditions[j] = appdto.ConditionView{
				ID:        c.ID,
				Attribute: c.Attribute,
				Operator:  c.Operator,
				Value:     c.Value,
			}
		}

		createdAt, _ := time.Parse(time.RFC3339, rg.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, rg.UpdatedAt)

		ruleGroups[i] = appdto.RuleGroupView{
			ID:         rg.ID,
			Name:       rg.Name,
			Variation:  rg.Variation,
			Priority:   rg.Priority,
			Conditions: conditions,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, v.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, v.UpdatedAt)

	return &appdto.FeatureFlagView{
		ID:                v.ID,
		Name:              v.Name,
		Key:               v.Key,
		Description:       v.Description,
		FlagType:          v.FlagType,
		DefaultValue:      v.DefaultValue,
		RolloutPercentage: v.RolloutPercentage,
		IsActive:          v.IsActive,
		RuleGroups:        ruleGroups,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}
