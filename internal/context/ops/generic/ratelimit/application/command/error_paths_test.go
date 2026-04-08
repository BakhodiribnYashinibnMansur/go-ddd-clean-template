package command

import (
	"context"
	"errors"
	"testing"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/outbox"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorRateLimitRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error)
}

func (m *errorRateLimitRepo) Save(_ context.Context, _ *ratelimitentity.RateLimit) error {
	return m.saveErr
}

func (m *errorRateLimitRepo) FindByID(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, ratelimitentity.ErrRateLimitNotFound
}

func (m *errorRateLimitRepo) Update(_ context.Context, _ *ratelimitentity.RateLimit) error {
	return m.updateErr
}

func (m *errorRateLimitRepo) Delete(_ context.Context, _ ratelimitentity.RateLimitID) error {
	return m.deleteErr
}

func (m *errorRateLimitRepo) List(_ context.Context, _ ratelimitrepo.RateLimitFilter) ([]*ratelimitentity.RateLimit, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreateRateLimitHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorRateLimitRepo{saveErr: errRepoSave}
	handler := NewCreateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), CreateRateLimitCommand{
		Name: "test", Rule: "/r", RequestsPerWindow: 10, WindowDuration: 30,
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateRateLimitHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	rl := ratelimitentity.NewRateLimit("n", "/r", 10, 30, true)

	repo := &errorRateLimitRepo{
		findFn:    func(_ context.Context, _ ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) { return rl, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRateLimitCommand{ID: ratelimitentity.RateLimitID(rl.ID())})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteRateLimitHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorRateLimitRepo{deleteErr: errRepoDelete}
	handler := NewDeleteRateLimitHandler(repo, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRateLimitCommand{ID: ratelimitentity.NewRateLimitID()})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}
