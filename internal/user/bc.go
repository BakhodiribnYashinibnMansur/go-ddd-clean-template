package user

import (
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/application/command"
	"gct/internal/user/application/query"
	"gct/internal/user/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the User BC.
type BoundedContext struct {
	// Commands
	CreateUser  *command.CreateUserHandler
	UpdateUser  *command.UpdateUserHandler
	DeleteUser  *command.DeleteUserHandler
	SignIn      *command.SignInHandler
	SignUp      *command.SignUpHandler
	SignOut     *command.SignOutHandler
	ApproveUser *command.ApproveUserHandler
	ChangeRole  *command.ChangeRoleHandler
	BulkAction  *command.BulkActionHandler

	// Queries
	GetUser   *query.GetUserHandler
	ListUsers *query.ListUsersHandler
}

// NewBoundedContext creates a fully wired User bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewUserWriteRepo(pool)
	readRepo := postgres.NewUserReadRepo(pool)

	return &BoundedContext{
		CreateUser:  command.NewCreateUserHandler(writeRepo, eventBus, l),
		UpdateUser:  command.NewUpdateUserHandler(writeRepo, eventBus, l),
		DeleteUser:  command.NewDeleteUserHandler(writeRepo, eventBus, l),
		SignIn:      command.NewSignInHandler(writeRepo, eventBus, l),
		SignUp:      command.NewSignUpHandler(writeRepo, eventBus, l),
		SignOut:     command.NewSignOutHandler(writeRepo, eventBus, l),
		ApproveUser: command.NewApproveUserHandler(writeRepo, eventBus, l),
		ChangeRole:  command.NewChangeRoleHandler(writeRepo, eventBus, l),
		BulkAction:  command.NewBulkActionHandler(writeRepo, eventBus, l),
		GetUser:     query.NewGetUserHandler(readRepo),
		ListUsers:   query.NewListUsersHandler(readRepo),
	}
}
