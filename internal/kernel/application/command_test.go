package application_test

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
)

// createUserCommand is a sample command for testing.
type createUserCommand struct {
	Name string
}

// mockCommandHandler is a mock that records whether Handle was called.
type mockCommandHandler struct {
	called bool
	cmd    createUserCommand
}

func (m *mockCommandHandler) Handle(ctx context.Context, cmd createUserCommand) error {
	m.called = true
	m.cmd = cmd
	return nil
}

// Compile-time interface satisfaction check.
var _ application.CommandHandler[createUserCommand] = (*mockCommandHandler)(nil)

func TestCommandHandler_Handle(t *testing.T) {
	handler := &mockCommandHandler{}
	cmd := createUserCommand{Name: "Alice"}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handler.called {
		t.Fatal("expected Handle to be called")
	}
	if handler.cmd.Name != "Alice" {
		t.Fatalf("expected command name Alice, got %s", handler.cmd.Name)
	}
}
