package client

import (
	"context"
	"testing"

	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"
	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"
	"gct/test/integration/common/setup"

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
	return user.NewBoundedContext(setup.TestPG.Pool, eb, outbox.NewEventCommitter(setup.TestPG.Pool, nil, eb, l), l, newTestJWTConfig(t))
}

func TestIntegration_CreateAndGetUser(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

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

	getResult, err := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userView.ID)})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if getResult.ID != userView.ID {
		t.Errorf("ID mismatch: %s vs %s", getResult.ID, userView.ID)
	}
}

func TestIntegration_UpdateUser(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateUser.Handle(ctx, command.CreateUserCommand{
		Phone:    "+998902222222",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID

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

	view, _ := bc.GetUser.Handle(ctx, query.GetUserQuery{ID: userentity.UserID(userID)})
	if view.Email == nil || *view.Email != "updated@example.com" {
		t.Error("email not updated")
	}
	if view.Username == nil || *view.Username != "updateduser" {
		t.Error("username not updated")
	}
}

func TestIntegration_DeleteUser(t *testing.T) {
	cleanDB(t)
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

	err = bc.DeleteUser.Handle(ctx, command.DeleteUserCommand{ID: userentity.UserID(userID)})
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	list2, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 users after delete, got %d", list2.Total)
	}
}

func TestIntegration_SignUp_SignIn_SignOut(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.SignUp.Handle(ctx, command.SignUpCommand{
		Phone:    "+998904444444",
		Password: "StrongP@ss123",
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	list, _ := bc.ListUsers.Handle(ctx, query.ListUsersQuery{
		Filter: userentity.UsersFilter{Pagination: &shared.Pagination{Limit: 10}},
	})
	userID := list.Users[0].ID
	_ = bc.ApproveUser.Handle(ctx, command.ApproveUserCommand{ID: userentity.UserID(userID)})

	result, err := bc.SignIn.Handle(ctx, command.SignInCommand{
		Login:      "+998904444444",
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

	err = bc.SignOut.Handle(ctx, command.SignOutCommand{
		UserID:    userentity.UserID(result.UserID),
		SessionID: userentity.SessionID(result.SessionID),
	})
	if err != nil {
		t.Fatalf("SignOut: %v", err)
	}
}

func TestIntegration_ChangeRole(t *testing.T) {
	cleanDB(t)
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

	var newRoleID uuid.UUID
	err = setup.TestPG.Pool.QueryRow(ctx, "SELECT id FROM role WHERE name = 'manager' LIMIT 1").Scan(&newRoleID)
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
