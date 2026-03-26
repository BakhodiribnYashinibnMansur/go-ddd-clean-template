package command

import (
	"context"
	"testing"
)

func TestDeleteScopeHandler_Success(t *testing.T) {
	deleted := false
	repo := &mockScopeRepository{
		deleteFn: func(_ context.Context, path, method string) error {
			if path != "/api/v1/users" {
				t.Errorf("expected path '/api/v1/users', got '%s'", path)
			}
			if method != "DELETE" {
				t.Errorf("expected method 'DELETE', got '%s'", method)
			}
			deleted = true
			return nil
		},
	}
	log := &mockLogger{}

	handler := NewDeleteScopeHandler(repo, log)

	cmd := DeleteScopeCommand{
		Path:   "/api/v1/users",
		Method: "DELETE",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !deleted {
		t.Error("expected delete to be called on repository")
	}
}
