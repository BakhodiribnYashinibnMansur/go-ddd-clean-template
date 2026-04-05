package query

import (
	"context"

	"gct/internal/context/admin/generic/featureflag/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// EvalResult is the value/type pair returned by the flag evaluator. Defined
// in the application layer so the FlagEvaluator port does not depend on the
// infrastructure cache package (DDD layering).
type EvalResult struct {
	Value    string
	FlagType string
}

// FlagEvaluator is the port the evaluate handler needs from the cache layer.
// Infrastructure adapters (cache) implement this interface.
type FlagEvaluator interface {
	EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *EvalResult
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
