package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"
	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Error-returning mocks
// ---------------------------------------------------------------------------

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errEventPublish = errors.New("event publish failed")

type errorRepo struct {
	mockUserRepository
	saveErr   error
	updateErr error
	findUser  *domain.User
}

func (m *errorRepo) Save(_ context.Context, _ *domain.User) error {
	return m.saveErr
}

func (m *errorRepo) Update(_ context.Context, _ *domain.User) error {
	return m.updateErr
}

func (m *errorRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findUser != nil && m.findUser.ID() == id {
		return m.findUser, nil
	}
	return nil, domain.ErrUserNotFound
}

type errorEventBus struct {
	err error
}

func (m *errorEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error {
	return m.err
}

func (m *errorEventBus) Subscribe(_ string, _ application.EventHandler) error {
	return nil
}

// ---------------------------------------------------------------------------
// Tests: Repository save errors
// ---------------------------------------------------------------------------

func TestCreateUserHandler_RepoSaveError(t *testing.T) {
	repo := &errorRepo{saveErr: errRepoSave}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), CreateUserCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	})
	if err == nil {
		t.Fatal("expected repo save error")
	}
	if !errors.Is(err, errRepoSave) {
		t.Errorf("expected errRepoSave, got: %v", err)
	}
}

func TestSignUpHandler_RepoSaveError(t *testing.T) {
	repo := &errorRepo{saveErr: errRepoSave}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), SignUpCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: Repository update errors
// ---------------------------------------------------------------------------

func TestUpdateUserHandler_RepoUpdateError(t *testing.T) {
	user := makeTestUser(t)
	repo := &errorRepo{
		updateErr: errRepoUpdate,
		findUser:  user,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	newName := "updated"
	err := handler.Handle(context.Background(), UpdateUserCommand{
		ID:       user.ID(),
		Username: &newName,
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestApproveUserHandler_RepoUpdateError(t *testing.T) {
	user := makeTestUser(t)
	repo := &errorRepo{
		updateErr: errRepoUpdate,
		findUser:  user,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewApproveUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), ApproveUserCommand{ID: user.ID()})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteUserHandler_RepoUpdateError(t *testing.T) {
	user := makeTestUser(t)
	repo := &errorRepo{
		updateErr: errRepoUpdate,
		findUser:  user,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), DeleteUserCommand{ID: user.ID()})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestChangeRoleHandler_RepoUpdateError(t *testing.T) {
	user := makeTestUser(t)
	repo := &errorRepo{
		updateErr: errRepoUpdate,
		findUser:  user,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewChangeRoleHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), ChangeRoleCommand{
		UserID: user.ID(),
		RoleID: uuid.New(),
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: Event publishing errors (should NOT propagate — logged only)
// ---------------------------------------------------------------------------

func TestCreateUserHandler_EventPublishError_StillSucceeds(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &errorEventBus{err: errEventPublish}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), CreateUserCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	})
	// Event publish errors are logged, not returned
	if err != nil {
		t.Fatalf("expected no error (event publish errors are logged), got: %v", err)
	}
	if repo.savedUser == nil {
		t.Fatal("user should still be saved")
	}
}

// ---------------------------------------------------------------------------
// Tests: Context cancellation
// ---------------------------------------------------------------------------

func TestCreateUserHandler_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateUserHandler(repo, eventBus, log)

	// The handler itself doesn't check ctx, but the repo.Save might.
	// With our mock, it will succeed. This tests that the handler passes ctx through.
	err := handler.Handle(ctx, CreateUserCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	})
	// Mock doesn't check context, so this will succeed.
	// The test verifies the handler doesn't panic with cancelled context.
	_ = err
}

// ---------------------------------------------------------------------------
// Tests: SignIn with repo update error
// ---------------------------------------------------------------------------

func TestSignInHandler_RepoUpdateError(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)
	user.Approve()

	repo := &signInErrorRepo{user: user, updateErr: errRepoUpdate}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignInHandler(repo, eventBus, log, testJWTConfig(t))

	_, err := handler.Handle(context.Background(), SignInCommand{
		Login:      "+998901234567",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "10.0.0.1",
		UserAgent:  "TestAgent",
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

// signInErrorRepo returns user by phone but fails on Update.
type signInErrorRepo struct {
	mockUserRepository
	user      *domain.User
	updateErr error
}

func (m *signInErrorRepo) FindByPhone(_ context.Context, phone domain.Phone) (*domain.User, error) {
	if m.user != nil && m.user.Phone().Value() == phone.Value() {
		return m.user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *signInErrorRepo) FindByEmail(_ context.Context, email domain.Email) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *signInErrorRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *signInErrorRepo) Update(_ context.Context, _ *domain.User) error {
	return m.updateErr
}
