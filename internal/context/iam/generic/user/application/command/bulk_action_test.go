package command

import (
	"context"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestBulkActionHandler_Activate(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	user.Deactivate()

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{user.TypedID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	_, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user.ID())},
		Action: BulkActionActivate,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.TypedID()]
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

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{user.TypedID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	_, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user.ID())},
		Action: BulkActionDeactivate,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.TypedID()]
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

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{user.TypedID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	_, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user.ID())},
		Action: BulkActionDelete,
	})
	require.NoError(t, err)

	updated := repo.updatedUsers[user.TypedID()]
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

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{user.TypedID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	_, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user.ID())},
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

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{
		user1.TypedID(): user1,
		user2.TypedID(): user2,
	}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	_, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user1.TypedID()), userentity.UserID(user2.TypedID())},
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

	repo := &bulkMockRepo{users: map[userentity.UserID]*userentity.User{user.TypedID(): user}}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewBulkActionHandler(repo, eventBus, log)

	result, err := handler.Handle(context.Background(), BulkActionCommand{
		IDs:    []userentity.UserID{userentity.UserID(user.ID()), userentity.NewUserID()}, // second ID doesn't exist
		Action: BulkActionActivate,
	})
	require.Error(t, err)
	require.Equal(t, 1, result.Succeeded)
	require.Equal(t, 1, result.Failed)

	if len(repo.updatedUsers) != 1 {
		t.Fatalf("expected 1 updated user, got %d", len(repo.updatedUsers))
	}
}

// bulkMockRepo supports multiple users for bulk action testing.
type bulkMockRepo struct {
	mockUserRepository
	users        map[userentity.UserID]*userentity.User
	updatedUsers map[userentity.UserID]*userentity.User
}

func (m *bulkMockRepo) FindByID(_ context.Context, id userentity.UserID) (*userentity.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *bulkMockRepo) Update(_ context.Context, entity *userentity.User) error {
	if m.updatedUsers == nil {
		m.updatedUsers = make(map[userentity.UserID]*userentity.User)
	}
	m.updatedUsers[entity.TypedID()] = entity
	return nil
}
