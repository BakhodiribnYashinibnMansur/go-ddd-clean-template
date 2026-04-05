package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBulkActionHandler_Activate(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	user.Deactivate()

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{user.ID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user.ID())},
		Action: BulkActionActivate,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.ID()]
	if updated == nil {
		t.Fatal("expected user to be updated")
	}
	if !updated.IsActive() {
		t.Error("expected user to be active after bulk activate")
	}
}

func TestBulkActionHandler_Deactivate(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{user.ID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user.ID())},
		Action: BulkActionDeactivate,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.ID()]
	if updated == nil {
		t.Fatal("expected user to be updated")
	}
	if updated.IsActive() {
		t.Error("expected user to be inactive after bulk deactivate")
	}
}

func TestBulkActionHandler_Delete(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{user.ID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user.ID())},
		Action: BulkActionDelete,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.ID()]
	if updated == nil {
		t.Fatal("expected user to be updated")
	}
	if updated.IsActive() {
		t.Error("expected user to be inactive after bulk delete")
	}
	if updated.DeletedAt() == nil {
		t.Error("expected deletedAt to be set after bulk delete")
	}
}

func TestBulkActionHandler_UnknownAction(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{user.ID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user.ID())},
		Action: "unknown_action",
	})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestBulkActionHandler_MultipleUsers(t *testing.T) {
	t.Parallel()

	user1 := makeTestUser(t)
	user2 := makeTestUser(t)

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{
		user1.ID(): user1,
		user2.ID(): user2,
	}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user1.ID()), domain.UserID(user2.ID())},
		Action: BulkActionDeactivate,
	})
	require.NoError(t, err)

	if len(repo.updatedUsers) != 2 {
		t.Fatalf("expected 2 updated users, got %d", len(repo.updatedUsers))
	}
}

func TestBulkActionHandler_SkipsMissing(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)

	repo := &bulkMockRepo{users: map[uuid.UUID]*domain.User{user.ID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []domain.UserID{domain.UserID(user.ID()), domain.NewUserID()}, // second ID doesn't exist
		Action: BulkActionActivate,
	})
	require.NoError(t, err)

	if len(repo.updatedUsers) != 1 {
		t.Fatalf("expected 1 updated user, got %d", len(repo.updatedUsers))
	}
}

// bulkMockRepo supports multiple users for bulk action testing.
type bulkMockRepo struct {
	mockUserRepository
	users        map[uuid.UUID]*domain.User
	updatedUsers map[uuid.UUID]*domain.User
}

func (m *bulkMockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *bulkMockRepo) Update(_ context.Context, entity *domain.User) error {
	if m.updatedUsers == nil {
		m.updatedUsers = make(map[uuid.UUID]*domain.User)
	}
	m.updatedUsers[entity.ID()] = entity
	return nil
}
