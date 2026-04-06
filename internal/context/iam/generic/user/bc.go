package user

import (
	"context"

	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	"gct/internal/context/iam/generic/user/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/revocation"

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
//
// maxSessionsFn is an optional per-user active-session cap resolver: when
// nil, the sign-in handler falls back to a constant default. In production
// this is wired from the SiteSetting BC so administrators can change the
// cap at runtime without a redeploy; in tests callers typically omit it.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, jwtCfg command.JWTConfig, maxSessionsFn ...func(ctx context.Context) int) *BoundedContext {
	writeRepo := postgres.NewUserWriteRepo(pool)
	readRepo := postgres.NewUserReadRepo(pool)

	return &BoundedContext{
		CreateUser:      command.NewCreateUserHandler(writeRepo, eventBus, l),
		UpdateUser:      command.NewUpdateUserHandler(writeRepo, eventBus, l),
		DeleteUser:      command.NewDeleteUserHandler(writeRepo, eventBus, l),
		SignIn:          command.NewSignInHandler(writeRepo, eventBus, l, jwtCfg, maxSessionsFn...),
		SignUp:          command.NewSignUpHandler(writeRepo, eventBus, l),
		SignOut:         command.NewSignOutHandler(writeRepo, eventBus, l),
		ApproveUser:     command.NewApproveUserHandler(writeRepo, eventBus, l),
		ChangeRole:      command.NewChangeRoleHandler(writeRepo, eventBus, l),
		BulkAction:      command.NewBulkActionHandler(writeRepo, eventBus, l),
		RevokeAll:       command.NewRevokeAllSessionsHandler(writeRepo, eventBus, l),
		GetUser:         query.NewGetUserHandler(readRepo, l),
		ListUsers:       query.NewListUsersHandler(readRepo, l),
		FindSession:     query.NewFindSessionHandler(readRepo, l),
		FindUserForAuth: query.NewFindUserForAuthHandler(readRepo, l),
	}
}

// WireSecurityDeps injects Phase S1 security dependencies into the sign-in
// and sign-out handlers. Call after NewBoundedContext; safe to omit entirely
// — all handlers degrade gracefully when deps are nil.
func (bc *BoundedContext) WireSecurityDeps(al audit.Logger, rs *revocation.Store, signInDeps command.SignInSecurityDeps) {
	bc.SignIn.WithSecurityDeps(signInDeps)
	bc.SignOut.WithSecurityDeps(al, rs)
}
