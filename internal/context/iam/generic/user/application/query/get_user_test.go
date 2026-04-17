package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock Read Repository ---

type mockUserReadRepository struct {
	view     *userentity.UserView
	views    []*userentity.UserView
	total    int64
	session  *shared.AuthSession
	authUser *shared.AuthUser
}

// errorReadRepo always returns an error.
type errorReadRepo struct {
	err error
}

func (m *errorReadRepo) FindByID(_ context.Context, _ userentity.UserID) (*userentity.UserView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ userentity.UsersFilter) ([]*userentity.UserView, int64, error) {
	return nil, 0, m.err
}

func (m *errorReadRepo) FindSessionByID(_ context.Context, _ userentity.SessionID) (*shared.AuthSession, error) {
	return nil, m.err
}

func (m *errorReadRepo) FindUserForAuth(_ context.Context, _ userentity.UserID) (*shared.AuthUser, error) {
	return nil, m.err
}

func (m *errorReadRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.Nil, m.err
}

var errRepoFailure = errors.New("repository failure")

func (m *mockUserReadRepository) FindByID(_ context.Context, id userentity.UserID) (*userentity.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserReadRepository) List(_ context.Context, _ userentity.UsersFilter) ([]*userentity.UserView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockUserReadRepository) FindSessionByID(_ context.Context, _ userentity.SessionID) (*shared.AuthSession, error) {
	if m.session != nil {
		return m.session, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserReadRepository) FindUserForAuth(_ context.Context, _ userentity.UserID) (*shared.AuthUser, error) {
	if m.authUser != nil {
		return m.authUser, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserReadRepository) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

// --- Tests ---

func TestGetUserHandler_Handle(t *testing.T) {
	t.Parallel()

	userID := userentity.NewUserID()
	phone := "+998901234567"
	email := "test@example.com"

	readRepo := &mockUserReadRepository{
		view: &userentity.UserView{
			ID:         userID,
			Phone:      phone,
			Email:      &email,
			Active:     true,
			IsApproved: true,
		},
	}

	handler := NewGetUserHandler(readRepo, logger.Noop())

	q := GetUserQuery{ID: userentity.UserID(userID)}
	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected user view, got nil")
	}

	if result.ID != uuid.UUID(userID) {
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

	q := GetUserQuery{ID: userentity.NewUserID()}
	_, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestGetUserHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	userID := userentity.NewUserID()
	roleID := uuid.New()
	phone := "+998901234567"
	email := "full@example.com"
	username := "fulluser"

	readRepo := &mockUserReadRepository{
		view: &userentity.UserView{
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
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: userentity.UserID(userID)})
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

	userID := userentity.NewUserID()

	readRepo := &mockUserReadRepository{
		view: &userentity.UserView{
			ID:     userID,
			Phone:  "+998900000000",
			Active: false,
		},
	}

	handler := NewGetUserHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetUserQuery{ID: userentity.UserID(userID)})
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
	_, err := handler.Handle(context.Background(), GetUserQuery{ID: userentity.NewUserID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
