package webhook

import (
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/webhook/application/command"
	"gct/internal/webhook/application/query"
	"gct/internal/webhook/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Webhook BC.
type BoundedContext struct {
	// Commands
	CreateWebhook *command.CreateHandler
	UpdateWebhook *command.UpdateHandler
	DeleteWebhook *command.DeleteHandler

	// Queries
	GetWebhook   *query.GetHandler
	ListWebhooks *query.ListHandler
}

// NewBoundedContext creates a fully wired Webhook bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewWebhookWriteRepo(pool)
	readRepo := postgres.NewWebhookReadRepo(pool)

	return &BoundedContext{
		CreateWebhook: command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateWebhook: command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteWebhook: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetWebhook:    query.NewGetHandler(readRepo),
		ListWebhooks:  query.NewListHandler(readRepo),
	}
}
