package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gct/internal/kernel/domain"
)

// mockEntity is a dummy entity for compile-time checks.
type mockEntity struct {
	ID   uuid.UUID
	Name string
}

// mockRepo implements domain.Repository[mockEntity] for compile-time satisfaction.
type mockRepo struct{}

func (m *mockRepo) Save(_ context.Context, _ domain.Querier, _ *mockEntity) error   { return nil }
func (m *mockRepo) FindByID(_ context.Context, _ uuid.UUID) (*mockEntity, error)    { return nil, nil }
func (m *mockRepo) Update(_ context.Context, _ domain.Querier, _ *mockEntity) error { return nil }
func (m *mockRepo) Delete(_ context.Context, _ domain.Querier, _ uuid.UUID) error   { return nil }

// Compile-time interface satisfaction check.
var _ domain.Repository[mockEntity, uuid.UUID] = (*mockRepo)(nil)

func TestRepository_CompileTimeCheck(t *testing.T) {
	// If this compiles, the interface is properly defined.
	t.Log("Repository[T] interface satisfaction verified")
}
