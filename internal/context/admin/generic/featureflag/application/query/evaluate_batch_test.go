package query_test

import (
	"context"
	"testing"

	"gct/internal/context/admin/generic/featureflag/application/query"

	"github.com/stretchr/testify/require"
)

type batchMockEvaluator struct {
	results map[string]*query.EvalResult
}

func (m *batchMockEvaluator) EvaluateFull(_ context.Context, key string, _ map[string]string) *query.EvalResult {
	return m.results[key]
}

func TestBatchEvaluateHandler_ReturnsMultiple(t *testing.T) {
	t.Parallel()

	eval := &batchMockEvaluator{results: map[string]*query.EvalResult{
		"flag_a": {Value: "true", FlagType: "bool"},
		"flag_b": {Value: "dark", FlagType: "string"},
	}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"flag_a", "flag_b", "flag_missing"},
		UserAttrs: map[string]string{"platform": "web"},
	})
	require.NoError(t, err)
	if len(result.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result.Flags))
	}
	if result.Flags["flag_a"].Value != "true" {
		t.Errorf("expected flag_a value true, got %s", result.Flags["flag_a"].Value)
	}
	if result.Flags["flag_b"].FlagType != "string" {
		t.Errorf("expected flag_b type string, got %s", result.Flags["flag_b"].FlagType)
	}
}

func TestBatchEvaluateHandler_AllMissing(t *testing.T) {
	t.Parallel()

	eval := &batchMockEvaluator{results: map[string]*query.EvalResult{}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"missing_a", "missing_b"},
		UserAttrs: nil,
	})
	require.NoError(t, err)
	if len(result.Flags) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(result.Flags))
	}
}
