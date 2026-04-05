package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/ops/ratelimit/domain"

	"github.com/google/uuid"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorRateLimitRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error)
}

func (m *errorRateLimitRepo) Save(_ context.Context, _ *domain.RateLimit) error {
	return m.saveErr
}

func (m *errorRateLimitRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrRateLimitNotFound
}

func (m *errorRateLimitRepo) Update(_ context.Context, _ *domain.RateLimit) error {
	return m.updateErr
}

func (m *errorRateLimitRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *errorRateLimitRepo) List(_ context.Context, _ domain.RateLimitFilter) ([]*domain.RateLimit, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreateRateLimitHandler_RepoError(t *testing.T) {
	repo := &errorRateLimitRepo{saveErr: errRepoSave}
	handler := NewCreateRateLimitHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateRateLimitCommand{
		Name: "test", Rule: "/r", RequestsPerWindow: 10, WindowDuration: 30,
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateRateLimitHandler_RepoUpdateError(t *testing.T) {
	rl := domain.NewRateLimit("n", "/r", 10, 30, true)

	repo := &errorRateLimitRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.RateLimit, error) { return rl, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateRateLimitHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRateLimitCommand{ID: rl.ID()})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteRateLimitHandler_RepoError(t *testing.T) {
	repo := &errorRateLimitRepo{deleteErr: errRepoDelete}
	handler := NewDeleteRateLimitHandler(repo, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRateLimitCommand{ID: uuid.New()})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}
