package command

import (
	"context"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// signUpMockReadRepo satisfies UserReadRepository for sign-up tests.
type signUpMockReadRepo struct{}

func (m *signUpMockReadRepo) FindByID(_ context.Context, _ userentity.UserID) (*userentity.UserView, error) {
	return nil, nil
}
func (m *signUpMockReadRepo) List(_ context.Context, _ userentity.UsersFilter) ([]*userentity.UserView, int64, error) {
	return nil, 0, nil
}
func (m *signUpMockReadRepo) FindSessionByID(_ context.Context, _ userentity.SessionID) (*shared.AuthSession, error) {
	return nil, nil
}
func (m *signUpMockReadRepo) FindUserForAuth(_ context.Context, _ userentity.UserID) (*shared.AuthUser, error) {
	return nil, nil
}
func (m *signUpMockReadRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

func TestSignUpHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, &signUpMockReadRepo{}, fakeDB{}, eventBus, log)

	username := "newuser"
	email := "newuser@example.com"
	cmd := SignUpCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
		Username: &username,
		Email:    &email,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved")
	}

	if repo.savedUser.Phone().Value() != "+998901234567" {
		t.Errorf("expected phone +998901234567, got %s", repo.savedUser.Phone().Value())
	}

	if !repo.savedUser.IsApproved() {
		t.Error("sign-up user should be auto-approved")
	}

	if repo.savedUser.Email() == nil || repo.savedUser.Email().Value() != "newuser@example.com" {
		t.Error("expected email to be set")
	}

	if repo.savedUser.Username() == nil || *repo.savedUser.Username() != "newuser" {
		t.Error("expected username to be set")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected events to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "user.created" {
		t.Errorf("expected user.created event, got %s", eventBus.publishedEvents[0].EventName())
	}
}

func TestSignUpHandler_MinimalFields(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, &signUpMockReadRepo{}, fakeDB{}, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "+998907654321",
		Password: "AnotherP@ss1",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved")
	}

	if repo.savedUser.Email() != nil {
		t.Error("email should be nil when not provided")
	}

	if repo.savedUser.Username() != nil {
		t.Error("username should be nil when not provided")
	}
}

func TestSignUpHandler_InvalidPhone(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, &signUpMockReadRepo{}, fakeDB{}, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "bad-phone",
		Password: "StrongP@ss123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}

	if repo.savedUser != nil {
		t.Error("no user should be saved for invalid phone")
	}
}

func TestSignUpHandler_WeakPassword(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, &signUpMockReadRepo{}, fakeDB{}, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "+998901234567",
		Password: "short",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for weak password")
	}
}
