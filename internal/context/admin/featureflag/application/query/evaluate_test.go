package query_test

import (
	"context"
	"testing"

	"gct/internal/context/admin/featureflag/application/query"
	"github.com/stretchr/testify/require"
)

type mockEvaluator struct {
	result *query.EvalResult
}

func (m *mockEvaluator) EvaluateFull(_ context.Context, _ string, _ map[string]string) *query.EvalResult {
	return m.result
}

func TestEvaluateHandler_ReturnsValue(t *testing.T) {
	t.Parallel()

	eval := &mockEvaluator{result: &query.EvalResult{Value: "true", FlagType: "bool"}}
	h := query.NewEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.EvaluateQuery{
		Key:       "dark_mode",
		UserAttrs: map[string]string{"platform": "web"},
	})
	require.NoError(t, err)
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
	t.Parallel()

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
