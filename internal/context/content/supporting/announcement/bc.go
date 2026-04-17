package announcement

import (
	"gct/internal/context/content/supporting/announcement/application/command"
	"gct/internal/context/content/supporting/announcement/application/query"
	"gct/internal/context/content/supporting/announcement/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Announcement BC.
type BoundedContext struct {
	// Commands
	CreateAnnouncement *command.CreateAnnouncementHandler
	UpdateAnnouncement *command.UpdateAnnouncementHandler
	DeleteAnnouncement *command.DeleteAnnouncementHandler

	// Queries
	GetAnnouncement   *query.GetAnnouncementHandler
	ListAnnouncements *query.ListAnnouncementsHandler
}

// NewBoundedContext creates a fully wired Announcement bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewAnnouncementWriteRepo(pool)
	readRepo := postgres.NewAnnouncementReadRepo(pool)

	return &BoundedContext{
		CreateAnnouncement: command.NewCreateAnnouncementHandler(writeRepo, committer, l),
		UpdateAnnouncement: command.NewUpdateAnnouncementHandler(writeRepo, committer, l),
		DeleteAnnouncement: command.NewDeleteAnnouncementHandler(writeRepo, committer, l),
		GetAnnouncement:    query.NewGetAnnouncementHandler(readRepo, l),
		ListAnnouncements:  query.NewListAnnouncementsHandler(readRepo, l),
	}
}
