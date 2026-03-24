package query

import (
	"context"
	"testing"

	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// --- Mock Read Repository ---

type mockUserReadRepository struct {
	view  *domain.UserView
	views []*domain.UserView
	total int64
}

func (m *mockUserReadRepository) FindByID(_ context.Context, id uuid.UUID) (*domain.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserReadRepository) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return m.views, m.total, nil
}

// --- Tests ---

func TestGetUserHandler_Handle(t *testing.T) {
	userID := uuid.New()
	phone := "+998901234567"
	email := "test@example.com"

	readRepo := &mockUserReadRepository{
		view: &domain.UserView{
			ID:         userID,
			Phone:      phone,
			Email:      &email,
			Active:     true,
			IsApproved: true,
		},
	}

	handler := NewGetUserHandler(readRepo)

	q := GetUserQuery{ID: userID}
	result, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected user view, got nil")
	}

	if result.ID != userID {
		t.Errorf("expected ID %s, got %s", userID, result.ID)
	}

	if result.Phone != phone {
		t.Errorf("expected phone %s, got %s", phone, result.Phone)
	}

	if result.Email == nil || *result.Email != email {
		t.Error("expected email to be set")
	}

	if !result.Active {
		t.Error("expected user to be active")
	}

	if !result.IsApproved {
		t.Error("expected user to be approved")
	}
}

func TestGetUserHandler_NotFound(t *testing.T) {
	readRepo := &mockUserReadRepository{}

	handler := NewGetUserHandler(readRepo)

	q := GetUserQuery{ID: uuid.New()}
	_, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}
