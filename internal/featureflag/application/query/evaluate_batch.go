package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
)

// BatchEvaluateQuery holds the input for evaluating multiple feature flags.
type BatchEvaluateQuery struct {
	Keys      []string
	UserAttrs map[string]string
}

// BatchEvaluateResult holds the output of a batch flag evaluation.
type BatchEvaluateResult struct {
	Flags map[string]EvaluateResult
}

// BatchEvaluateHandler handles the BatchEvaluateQuery.
type BatchEvaluateHandler struct {
	evaluator FlagEvaluator
}

// NewBatchEvaluateHandler creates a new BatchEvaluateHandler.
func NewBatchEvaluateHandler(evaluator FlagEvaluator) *BatchEvaluateHandler {
	return &BatchEvaluateHandler{evaluator: evaluator}
}

// Handle evaluates multiple feature flags for the given user attributes.
// Flags that do not exist are omitted from the result.
func (h *BatchEvaluateHandler) Handle(ctx context.Context, q BatchEvaluateQuery) (_ *BatchEvaluateResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "BatchEvaluateHandler.Handle")
	defer func() { end(err) }()

	flags := make(map[string]EvaluateResult, len(q.Keys))
	for _, key := range q.Keys {
		result := h.evaluator.EvaluateFull(ctx, key, q.UserAttrs)
		if result == nil {
			continue
		}
		flags[key] = EvaluateResult{
			Key:      key,
			Value:    result.Value,
			FlagType: result.FlagType,
		}
	}

	return &BatchEvaluateResult{Flags: flags}, nil
}
