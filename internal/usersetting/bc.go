package usersetting

import (
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/usersetting/application/command"
	"gct/internal/usersetting/application/query"
	"gct/internal/usersetting/infrastructure/postgres"

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
		GetUserSetting:    query.NewGetUserSettingHandler(readRepo),
		ListUserSettings:  query.NewListUserSettingsHandler(readRepo),
	}
}
