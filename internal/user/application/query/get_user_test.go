package query

import (
	"context"
	"errors"
	"testing"

	shared "gct/internal/shared/domain"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// --- Mock Read Repository ---

type mockUserReadRepository struct {
	view  *domain.UserView
	views []*domain.UserView
	total int64
}

// errorReadRepo always returns an error.
type errorReadRepo struct {
	err error
}

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.UserView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return nil, 0, m.err
}

func (m *errorReadRepo) FindSessionByID(_ context.Context, _ uuid.UUID) (*shared.AuthSession, error) {
	return nil, m.err
}

func (m *errorReadRepo) FindUserForAuth(_ context.Context, _ uuid.UUID) (*shared.AuthUser, error) {
	return nil, m.err
}

var errRepoFailure = errors.New("repository failure")

func (m *mockUserReadRepository) FindByID(_ context.Context, id uuid.UUID) (*domain.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserReadRepository) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockUserReadRepository) FindSessionByID(_ context.Context, _ uuid.UUID) (*shared.AuthSession, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserReadRepository) FindUserForAuth(_ context.Context, _ uuid.UUID) (*shared.AuthUser, error) {
	return nil, domain.ErrUserNotFound
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

func TestGetUserHandler_AllFieldsMapped(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	phone := "+998901234567"
	email := "full@example.com"
	username := "fulluser"

	readRepo := &mockUserReadRepository{
		view: &domain.UserView{
			ID:         userID,
			Phone:      phone,
			Email:      &email,
			Username:   &username,
			RoleID:     &roleID,
			Attributes: map[string]any{"level": 5},
			Active:     true,
			IsApproved: true,
		},
	}

	handler := NewGetUserHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: userID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RoleID == nil || *result.RoleID != roleID {
		t.Error("roleID not mapped")
	}
	if result.Username == nil || *result.Username != "fulluser" {
		t.Error("username not mapped")
	}
	if result.Attributes["level"] != 5 {
		t.Error("attributes not mapped")
	}
}

func TestGetUserHandler_NilOptionalFields(t *testing.T) {
	userID := uuid.New()

	readRepo := &mockUserReadRepository{
		view: &domain.UserView{
			ID:     userID,
			Phone:  "+998900000000",
			Active: false,
		},
	}

	handler := NewGetUserHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: userID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Email != nil {
		t.Error("email should be nil")
	}
	if result.Username != nil {
		t.Error("username should be nil")
	}
	if result.RoleID != nil {
		t.Error("roleID should be nil")
	}
	if result.Active {
		t.Error("expected inactive user")
	}
}

func TestGetUserHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewGetUserHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetUserQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
