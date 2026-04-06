package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"time"

	"gct/internal/context/admin/generic/featureflag/application/dto"
	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single feature flag.
type GetQuery struct {
	ID ffentity.FeatureFlagID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo ffrepo.FeatureFlagReadRepository
	logger   logger.Log
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo ffrepo.FeatureFlagReadRepository, l logger.Log) *GetHandler {
	return &GetHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetQuery and returns a FeatureFlagView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *dto.FeatureFlagView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetFeatureFlag", "feature_flag")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "Get", Entity: "feature_flag", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return mapToAppView(view), nil
}

// mapToAppView converts a repository FeatureFlagView to an application FeatureFlagView.
func mapToAppView(v *ffrepo.FeatureFlagView) *dto.FeatureFlagView {
	ruleGroups := make([]dto.RuleGroupView, len(v.RuleGroups))
	for i, rg := range v.RuleGroups {
		conditions := make([]dto.ConditionView, len(rg.Conditions))
		for j, c := range rg.Conditions {
			conditions[j] = dto.ConditionView{
				ID:        c.ID,
				Attribute: c.Attribute,
				Operator:  c.Operator,
				Value:     c.Value,
			}
		}

		createdAt, _ := time.Parse(time.RFC3339, rg.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, rg.UpdatedAt)

		ruleGroups[i] = dto.RuleGroupView{
			ID:         uuid.UUID(rg.ID),
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

	return &dto.FeatureFlagView{
		ID:                uuid.UUID(v.ID),
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
