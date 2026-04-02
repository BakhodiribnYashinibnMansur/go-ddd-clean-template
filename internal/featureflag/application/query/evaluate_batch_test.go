package query_test

import (
	"context"
	"testing"

	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/infrastructure/cache"
)

type batchMockEvaluator struct {
	results map[string]*cache.EvalResult
}

func (m *batchMockEvaluator) EvaluateFull(_ context.Context, key string, _ map[string]string) *cache.EvalResult {
	return m.results[key]
}

func TestBatchEvaluateHandler_ReturnsMultiple(t *testing.T) {
	eval := &batchMockEvaluator{results: map[string]*cache.EvalResult{
		"flag_a": {Value: "true", FlagType: "bool"},
		"flag_b": {Value: "dark", FlagType: "string"},
	}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"flag_a", "flag_b", "flag_missing"},
		UserAttrs: map[string]string{"platform": "web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	eval := &batchMockEvaluator{results: map[string]*cache.EvalResult{}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"missing_a", "missing_b"},
		UserAttrs: nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Flags) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(result.Flags))
	}
}
