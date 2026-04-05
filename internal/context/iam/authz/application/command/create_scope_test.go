package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/domain"
	shared "gct/internal/platform/domain"
)

// --- Mock ScopeRepository ---

type mockScopeRepository struct {
	savedScope *domain.Scope
	saveFn     func(ctx context.Context, scope domain.Scope) error
	deleteFn   func(ctx context.Context, path, method string) error
}

func (m *mockScopeRepository) Save(ctx context.Context, scope domain.Scope) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, scope)
	}
	m.savedScope = &scope
	return nil
}

func (m *mockScopeRepository) Delete(ctx context.Context, path, method string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, path, method)
	}
	return nil
}

func (m *mockScopeRepository) List(ctx context.Context, pagination shared.Pagination) ([]domain.Scope, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreateScopeHandler_Success(t *testing.T) {
	repo := &mockScopeRepository{}
	log := &mockLogger{}

	handler := NewCreateScopeHandler(repo, log)

	cmd := CreateScopeCommand{
		Path:   "/api/v1/users",
		Method: "GET",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.savedScope == nil {
		t.Fatal("expected scope to be saved")
	}

	if repo.savedScope.Path != "/api/v1/users" {
		t.Errorf("expected path '/api/v1/users', got '%s'", repo.savedScope.Path)
	}

	if repo.savedScope.Method != "GET" {
		t.Errorf("expected method 'GET', got '%s'", repo.savedScope.Method)
	}
}
