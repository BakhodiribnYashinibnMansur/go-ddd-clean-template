package application_test

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
)

// getUserQuery is a sample query for testing.
type getUserQuery struct {
	ID string
}

// userResult is a sample result for testing.
type userResult struct {
	Name string
}

// mockQueryHandler is a mock that returns a predefined result.
type mockQueryHandler struct {
	result userResult
}

func (m *mockQueryHandler) Handle(_ context.Context, _ getUserQuery) (userResult, error) {
	return m.result, nil
}

// Compile-time interface satisfaction check.
var _ application.QueryHandler[getUserQuery, userResult] = (*mockQueryHandler)(nil)

func TestQueryHandler_Handle(t *testing.T) {
	handler := &mockQueryHandler{result: userResult{Name: "Bob"}}

	result, err := handler.Handle(context.Background(), getUserQuery{ID: "123"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "Bob" {
		t.Fatalf("expected result name Bob, got %s", result.Name)
	}
}
