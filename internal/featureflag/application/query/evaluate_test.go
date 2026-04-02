package query_test

import (
	"context"
	"testing"

	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/infrastructure/cache"
)

type mockEvaluator struct {
	result *cache.EvalResult
}

func (m *mockEvaluator) EvaluateFull(_ context.Context, _ string, _ map[string]string) *cache.EvalResult {
	return m.result
}

func TestEvaluateHandler_ReturnsValue(t *testing.T) {
	eval := &mockEvaluator{result: &cache.EvalResult{Value: "true", FlagType: "bool"}}
	h := query.NewEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.EvaluateQuery{
		Key:       "dark_mode",
		UserAttrs: map[string]string{"platform": "web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Key != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", result.Key)
	}
	if result.Value != "true" {
		t.Errorf("expected value true, got %s", result.Value)
	}
	if result.FlagType != "bool" {
		t.Errorf("expected flag_type bool, got %s", result.FlagType)
	}
}

func TestEvaluateHandler_FlagNotFound(t *testing.T) {
	eval := &mockEvaluator{result: nil}
	h := query.NewEvaluateHandler(eval)

	_, err := h.Handle(context.Background(), query.EvaluateQuery{
		Key:       "nonexistent",
		UserAttrs: nil,
	})
	if err == nil {
		t.Fatal("expected error for missing flag")
	}
}
