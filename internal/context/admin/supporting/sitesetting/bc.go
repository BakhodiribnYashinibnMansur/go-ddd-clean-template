package sitesetting

import (
	"gct/internal/context/admin/supporting/sitesetting/application/command"
	"gct/internal/context/admin/supporting/sitesetting/application/query"
	"gct/internal/context/admin/supporting/sitesetting/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the SiteSetting BC.
type BoundedContext struct {
	// Commands
	CreateSiteSetting *command.CreateSiteSettingHandler
	UpdateSiteSetting *command.UpdateSiteSettingHandler
	DeleteSiteSetting *command.DeleteSiteSettingHandler

	// Queries
	GetSiteSetting   *query.GetSiteSettingHandler
	ListSiteSettings *query.ListSiteSettingsHandler
	UserMaxSessions  *query.GetUserMaxSessionsHandler
}

// NewBoundedContext creates a fully wired SiteSetting bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewSiteSettingWriteRepo(pool)
	readRepo := postgres.NewSiteSettingReadRepo(pool)

	return &BoundedContext{
		CreateSiteSetting: command.NewCreateSiteSettingHandler(writeRepo, committer, l),
		UpdateSiteSetting: command.NewUpdateSiteSettingHandler(writeRepo, committer, l),
		DeleteSiteSetting: command.NewDeleteSiteSettingHandler(writeRepo, l),
		GetSiteSetting:    query.NewGetSiteSettingHandler(readRepo, l),
		ListSiteSettings:  query.NewListSiteSettingsHandler(readRepo, l),
		UserMaxSessions:   query.NewGetUserMaxSessionsHandler(readRepo, l),
	}
}
