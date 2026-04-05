package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	"gct/internal/context/iam/generic/user/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock Read Repository ---

type mockUserReadRepository struct {
	view     *domain.UserView
	views    []*domain.UserView
	total    int64
	session  *shared.AuthSession
	authUser *shared.AuthUser
}

// errorReadRepo always returns an error.
type errorReadRepo struct {
	err error
}

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.UserID) (*domain.UserView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return nil, 0, m.err
}

func (m *errorReadRepo) FindSessionByID(_ context.Context, _ domain.SessionID) (*shared.AuthSession, error) {
	return nil, m.err
}

func (m *errorReadRepo) FindUserForAuth(_ context.Context, _ domain.UserID) (*shared.AuthUser, error) {
	return nil, m.err
}

var errRepoFailure = errors.New("repository failure")

func (m *mockUserReadRepository) FindByID(_ context.Context, id domain.UserID) (*domain.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserReadRepository) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockUserReadRepository) FindSessionByID(_ context.Context, _ domain.SessionID) (*shared.AuthSession, error) {
	if m.session != nil {
		return m.session, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserReadRepository) FindUserForAuth(_ context.Context, _ domain.UserID) (*shared.AuthUser, error) {
	if m.authUser != nil {
		return m.authUser, nil
	}
	return nil, domain.ErrUserNotFound
}

// --- Tests ---

func TestGetUserHandler_Handle(t *testing.T) {
	t.Parallel()

	userID := domain.NewUserID()
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

	handler := NewGetUserHandler(readRepo, logger.Noop())

	q := GetUserQuery{ID: domain.UserID(userID)}
	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)

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
	t.Parallel()

	readRepo := &mockUserReadRepository{}

	handler := NewGetUserHandler(readRepo, logger.Noop())

	q := GetUserQuery{ID: domain.NewUserID()}
	_, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestGetUserHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	userID := domain.NewUserID()
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
			Attributes: map[string]string{"level": "5"},
			Active:     true,
			IsApproved: true,
		},
	}

	handler := NewGetUserHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: domain.UserID(userID)})
	require.NoError(t, err)

	if result.RoleID == nil || *result.RoleID != roleID {
		t.Error("roleID not mapped")
	}
	if result.Username == nil || *result.Username != "fulluser" {
		t.Error("username not mapped")
	}
	if result.Attributes["level"] != "5" {
		t.Error("attributes not mapped")
	}
}

func TestGetUserHandler_NilOptionalFields(t *testing.T) {
	t.Parallel()

	userID := domain.NewUserID()

	readRepo := &mockUserReadRepository{
		view: &domain.UserView{
			ID:     userID,
			Phone:  "+998900000000",
			Active: false,
		},
	}

	handler := NewGetUserHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: domain.UserID(userID)})
	require.NoError(t, err)

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
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewGetUserHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserQuery{ID: domain.NewUserID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
