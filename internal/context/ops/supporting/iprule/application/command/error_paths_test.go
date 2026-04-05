package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/ops/supporting/iprule/domain"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorIPRuleRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id domain.IPRuleID) (*domain.IPRule, error)
}

func (m *errorIPRuleRepo) Save(_ context.Context, _ *domain.IPRule) error {
	return m.saveErr
}

func (m *errorIPRuleRepo) FindByID(ctx context.Context, id domain.IPRuleID) (*domain.IPRule, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrIPRuleNotFound
}

func (m *errorIPRuleRepo) Update(_ context.Context, _ *domain.IPRule) error {
	return m.updateErr
}

func (m *errorIPRuleRepo) Delete(_ context.Context, _ domain.IPRuleID) error {
	return m.deleteErr
}

func (m *errorIPRuleRepo) List(_ context.Context, _ domain.IPRuleFilter) ([]*domain.IPRule, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreateIPRuleHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorIPRuleRepo{saveErr: errRepoSave}
	handler := NewCreateIPRuleHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateIPRuleCommand{
		IPAddress: "1.1.1.1", Action: "DENY", Reason: "test",
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateIPRuleHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	r := domain.NewIPRule("1.1.1.1", "DENY", "test", nil)

	repo := &errorIPRuleRepo{
		findFn:    func(_ context.Context, _ domain.IPRuleID) (*domain.IPRule, error) { return r, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateIPRuleHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateIPRuleCommand{ID: r.TypedID()})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteIPRuleHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorIPRuleRepo{deleteErr: errRepoDelete}
	handler := NewDeleteIPRuleHandler(repo, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteIPRuleCommand{ID: domain.NewIPRuleID()})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}
