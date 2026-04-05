package usersetting

import (
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/iam/generic/usersetting/application/command"
	"gct/internal/context/iam/generic/usersetting/application/query"
	"gct/internal/context/iam/generic/usersetting/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the UserSetting BC.
type BoundedContext struct {
	// Commands
	UpsertUserSetting *command.UpsertUserSettingHandler
	DeleteUserSetting *command.DeleteUserSettingHandler

	// Queries
	GetUserSetting   *query.GetUserSettingHandler
	ListUserSettings *query.ListUserSettingsHandler
}

// NewBoundedContext creates a fully wired UserSetting bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewUserSettingWriteRepo(pool)
	readRepo := postgres.NewUserSettingReadRepo(pool)

	return &BoundedContext{
		UpsertUserSetting: command.NewUpsertUserSettingHandler(writeRepo, eventBus, l),
		DeleteUserSetting: command.NewDeleteUserSettingHandler(writeRepo, l),
		GetUserSetting:    query.NewGetUserSettingHandler(readRepo, l),
		ListUserSettings:  query.NewListUserSettingsHandler(readRepo, l),
	}
}
