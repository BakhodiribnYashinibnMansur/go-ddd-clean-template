package notification

import (
	"gct/internal/context/content/generic/notification/application/command"
	"gct/internal/context/content/generic/notification/application/query"
	"gct/internal/context/content/generic/notification/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Notification BC.
type BoundedContext struct {
	// Commands
	CreateNotification *command.CreateHandler
	DeleteNotification *command.DeleteHandler

	// Queries
	GetNotification   *query.GetHandler
	ListNotifications *query.ListHandler
}

// NewBoundedContext creates a fully wired Notification bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewNotificationWriteRepo(pool)
	readRepo := postgres.NewNotificationReadRepo(pool)

	return &BoundedContext{
		CreateNotification: command.NewCreateHandler(writeRepo, committer, l),
		DeleteNotification: command.NewDeleteHandler(writeRepo, committer, l),
		GetNotification:    query.NewGetHandler(readRepo, l),
		ListNotifications:  query.NewListHandler(readRepo, l),
	}
}
