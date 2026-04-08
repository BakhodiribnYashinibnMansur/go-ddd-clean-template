package ddd

import (
	"context"
	"testing"

	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"
	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
)

func newTestJWTConfig(t *testing.T) command.JWTConfig {
	t.Helper()
	return command.JWTConfig{
		Issuer: "gct-test",
	}
}

func newTestBC(t *testing.T) *user.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return user.NewBoundedContext(testPool, eb, outbox.NewEventCommitter(testPool, nil, eb, l), l, newTestJWTConfig(t))
}

// ---------------------------------------------------------------------------
// Integration: CreateUser → GetUser → ListUsers
// ---------------------------------------------------------------------------

func TestIntegration_CreateAndGetUser(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create user
	email := "integration@example.com"
	username := "intuser"
	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998901111111",
		Password: "StrongP@ss123",
		Email:    &email,
		Username: &username,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// List users to get the ID
	result, err := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{
			Pagination: &shared.Pagination{Limit: 10, Offset: 0},
		},
	})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 user, got %d", result.Total)
	}

	userView := result.Users[0]
	if userView.Phone != "+998901111111" {
		t.Errorf("expected phone +998901111111, got %s", userView.Phone)
	}
	if userView.Email == nil || *userView.Email != "integration@example.com" {
		t.Error("expected email integration@example.com")
	}
	if userView.Username == nil || *userView.Username != "intuser" {
		t.Error("expected username intuser")
	}

	// Get user by ID
	getResult, err := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userView.ID)})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if getResult.ID != userView.ID {
		t.Errorf("ID mismatch: %s vs %s", getResult.ID, userView.ID)
	}
}

// ---------------------------------------------------------------------------
// Integration: UpdateUser
// ---------------------------------------------------------------------------

func TestIntegration_UpdateUser(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create
	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998902222222",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Get ID
	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID

	// Update
	newEmail := "updated@example.com"
	newName := "updateduser"
	err = bc.UpdateUser.Handle(ctx, command.UpdateUserCommand{
		ID:       userentity.UserID(userID),
		Email:    &newEmail,
		Username: &newName,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	// Verify
	view, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userID)})
	if view.Email == nil || *view.Email != "updated@example.com" {
		t.Error("email not updated")
	}
	if view.Username == nil || *view.Username != "updateduser" {
		t.Error("username not updated")
	}
}

// ---------------------------------------------------------------------------
// Integration: ApproveUser
// ---------------------------------------------------------------------------

func TestIntegration_ApproveUser(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998903333333",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID

	if list.Users[0].IsApproved {
		t.Fatal("new user should not be approved")
	}

	err = bc.ApproveUser.Handle(ctx, command.ApproveUserCommand{ID: userentity.UserID(userID)})
	if err != nil {
		t.Fatalf("ApproveUser: %v", err)
	}

	view, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userID)})
	if !view.IsApproved {
		t.Error("user should be approved")
	}
}

// ---------------------------------------------------------------------------
// Integration: ChangeRole
// ---------------------------------------------------------------------------

func TestIntegration_ChangeRole(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998904444444",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID

	var newRoleID uuid.UUID
	err = testPool.QueryRow(ctx, "SELECT id FROM role WHERE name = 'manager' LIMIT 1").Scan(&newRoleID)
	if err != nil {
		t.Fatalf("fetch role: %v", err)
	}

	err = bc.ChangeRole.Handle(ctx, command.ChangeRoleCommand{
		UserID: userentity.UserID(userID),
		RoleID: newRoleID,
	})
	if err != nil {
		t.Fatalf("ChangeRole: %v", err)
	}

	view, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userID)})
	if view.RoleID == nil || *view.RoleID != newRoleID {
		t.Error("role ID not updated")
	}
}

// ---------------------------------------------------------------------------
// Integration: DeleteUser (soft-delete)
// ---------------------------------------------------------------------------

func TestIntegration_DeleteUser(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998905555555",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID

	err = bc.DeleteUser.Handle(ctx, command.DeleteUserCommand{ID: userentity.UserID(userID)})
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	// Soft-deleted user should not appear in list
	list2, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 users after delete, got %d", list2.Total)
	}
}

// ---------------------------------------------------------------------------
// Integration: SignUp → SignIn → SignOut
// ---------------------------------------------------------------------------

func TestIntegration_SignUp_SignIn_SignOut(t *testing.T) {
	cleanUserTables(t)

	// Use event bus that tracks events
	eb := eventbus.NewInMemoryEventBus()
	var receivedEvents []string
	eb.Subscribe("user.created", func(_ context.Context, e shared.DomainEvent) error {
		receivedEvents = append(receivedEvents, e.EventName())
		return nil
	})
	eb.Subscribe("user.signed_in", func(_ context.Context, e shared.DomainEvent) error {
		receivedEvents = append(receivedEvents, e.EventName())
		return nil
	})

	l := logger.New("error")
	bc := user.NewBoundedContext(testPool, eb, outbox.NewEventCommitter(testPool, nil, eb, l), l, newTestJWTConfig(t))
	ctx := context.Background()

	// Sign Up
	err := bc.SignUp.Handle(ctx, command.SignUpCommand{
		Phone:    "+998906666666",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	// Verify event
	if len(receivedEvents) == 0 || receivedEvents[0] != "user.created" {
		t.Errorf("expected user.created event, got %v", receivedEvents)
	}

	// Approve user first (required for sign-in)
	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID
	bc.ApproveUser.Handle(ctx, command.ApproveUserCommand{ID: userentity.UserID(userID)})

	// Sign In
	result, err := bc.SignIn.Handle(ctx, command.SignInCommand{
		Login:      "+998906666666",
		Password:   "StrongP@ss123",
		DeviceType: "desktop",
		IP:         "10.0.0.1",
		UserAgent:  "IntegrationTest/1.0",
	})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if result.UserID != userID {
		t.Errorf("user ID mismatch: %s vs %s", result.UserID, userID)
	}

	// Sign Out
	err = bc.SignOut.Handle(ctx, command.SignOutCommand{
		UserID:    userentity.UserID(result.UserID),
		SessionID: userentity.SessionID(result.SessionID),
	})
	if err != nil {
		t.Fatalf("SignOut: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Integration: BulkAction
// ---------------------------------------------------------------------------

func TestIntegration_BulkAction(t *testing.T) {
	cleanUserTables(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create 3 users
	phones := []string{"+998907771111", "+998907772222", "+998907773333"}
	for _, phone := range phones {
		err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
			Phone:    phone,
			Password: "StrongP@ss123",
		})
		if err != nil {
			t.Fatalf("CreateUser(%s): %v", phone, err)
		}
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	if list.Total != 3 {
		t.Fatalf("expected 3 users, got %d", list.Total)
	}

	// Bulk deactivate first 2
	ids := []userentity.UserID{userentity.UserID(list.Users[0].ID), userentity.UserID(list.Users[1].ID)}
	err := bc.BulkAction.Handle(ctx, command.BulkActionCommand{
		IDs:    ids,
		Action: "deactivate",
	})
	if err != nil {
		t.Fatalf("BulkAction deactivate: %v", err)
	}

	// Verify: first 2 should be inactive
	for _, id := range ids {
		view, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: id})
		if view.Active {
			t.Errorf("user %s should be inactive", id)
		}
	}

	// Third should still be active
	view3, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(list.Users[2].ID)})
	if !view3.Active {
		t.Error("third user should still be active")
	}
}

// ---------------------------------------------------------------------------
// Integration: Event Bus subscription verification
// ---------------------------------------------------------------------------

func TestIntegration_EventBus_PublishesOnCreate(t *testing.T) {
	cleanUserTables(t)

	eb := eventbus.NewInMemoryEventBus()
	var events []shared.DomainEvent
	eb.Subscribe("user.created", func(_ context.Context, e shared.DomainEvent) error {
		events = append(events, e)
		return nil
	})

	l := logger.New("error")
	bc := user.NewBoundedContext(testPool, eb, outbox.NewEventCommitter(testPool, nil, eb, l), l, newTestJWTConfig(t))
	ctx := context.Background()

	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998908888888",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "user.created" {
		t.Errorf("expected user.created, got %s", events[0].EventName())
	}
}

// Ensure mockEventBus satisfies application.EventBus to keep compiler happy in other test files.
var _ application.EventBus = (*eventbus.InMemoryEventBus)(nil)
