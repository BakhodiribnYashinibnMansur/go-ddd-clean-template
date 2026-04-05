package user

import (
	"gct/internal/platform/application"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/context/iam/user/application/command"
	"gct/internal/context/iam/user/application/query"
	"gct/internal/context/iam/user/infrastructure/postgres"

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
	RevokeAll   *command.RevokeAllSessionsHandler

	// Queries
	GetUser         *query.GetUserHandler
	ListUsers       *query.ListUsersHandler
	FindSession     *query.FindSessionHandler
	FindUserForAuth *query.FindUserForAuthHandler
}

// NewBoundedContext creates a fully wired User bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, jwtCfg command.JWTConfig) *BoundedContext {
	writeRepo := postgres.NewUserWriteRepo(pool)
	readRepo := postgres.NewUserReadRepo(pool)

	return &BoundedContext{
		CreateUser:  command.NewCreateUserHandler(writeRepo, eventBus, l),
		UpdateUser:  command.NewUpdateUserHandler(writeRepo, eventBus, l),
		DeleteUser:  command.NewDeleteUserHandler(writeRepo, eventBus, l),
		SignIn:      command.NewSignInHandler(writeRepo, eventBus, l, jwtCfg),
		SignUp:      command.NewSignUpHandler(writeRepo, eventBus, l),
		SignOut:     command.NewSignOutHandler(writeRepo, eventBus, l),
		ApproveUser: command.NewApproveUserHandler(writeRepo, eventBus, l),
		ChangeRole:  command.NewChangeRoleHandler(writeRepo, eventBus, l),
		BulkAction:  command.NewBulkActionHandler(writeRepo, eventBus, l),
		RevokeAll:   command.NewRevokeAllSessionsHandler(writeRepo, eventBus, l),
		GetUser:         query.NewGetUserHandler(readRepo, l),
		ListUsers:       query.NewListUsersHandler(readRepo, l),
		FindSession:     query.NewFindSessionHandler(readRepo, l),
		FindUserForAuth: query.NewFindUserForAuthHandler(readRepo, l),
	}
}
