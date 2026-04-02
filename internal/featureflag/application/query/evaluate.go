package query

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/featureflag/infrastructure/cache"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"
)

// FlagEvaluator is the interface the evaluate handler needs from the cache layer.
type FlagEvaluator interface {
	EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *cache.EvalResult
}

// EvaluateQuery holds the input for evaluating a single feature flag.
type EvaluateQuery struct {
	Key       string
	UserAttrs map[string]string
}

// EvaluateResult holds the output of a single flag evaluation.
type EvaluateResult struct {
	Key      string
	Value    string
	FlagType string
}

// EvaluateHandler handles the EvaluateQuery.
type EvaluateHandler struct {
	evaluator FlagEvaluator
}

// NewEvaluateHandler creates a new EvaluateHandler.
func NewEvaluateHandler(evaluator FlagEvaluator) *EvaluateHandler {
	return &EvaluateHandler{evaluator: evaluator}
}

// Handle evaluates a single feature flag for the given user attributes.
func (h *EvaluateHandler) Handle(ctx context.Context, q EvaluateQuery) (_ *EvaluateResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "EvaluateHandler.Handle")
	defer func() { end(err) }()

	result := h.evaluator.EvaluateFull(ctx, q.Key, q.UserAttrs)
	if result == nil {
		return nil, apperrors.MapToServiceError(domain.ErrFeatureFlagNotFound)
	}

	return &EvaluateResult{
		Key:      q.Key,
		Value:    result.Value,
		FlagType: result.FlagType,
	}, nil
}
