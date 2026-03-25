package notification

import (
	"gct/internal/notification/application/command"
	"gct/internal/notification/application/query"
	"gct/internal/notification/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

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
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewNotificationWriteRepo(pool)
	readRepo := postgres.NewNotificationReadRepo(pool)

	return &BoundedContext{
		CreateNotification: command.NewCreateHandler(writeRepo, eventBus, l),
		DeleteNotification: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetNotification:    query.NewGetHandler(readRepo),
		ListNotifications:  query.NewListHandler(readRepo),
	}
}
