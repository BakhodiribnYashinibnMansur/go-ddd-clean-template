package emailtemplate

import (
	"gct/internal/emailtemplate/application/command"
	"gct/internal/emailtemplate/application/query"
	"gct/internal/emailtemplate/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the EmailTemplate BC.
type BoundedContext struct {
	// Commands
	CreateTemplate *command.CreateHandler
	UpdateTemplate *command.UpdateHandler
	DeleteTemplate *command.DeleteHandler

	// Queries
	GetTemplate   *query.GetHandler
	ListTemplates *query.ListHandler
}

// NewBoundedContext creates a fully wired EmailTemplate bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewEmailTemplateWriteRepo(pool)
	readRepo := postgres.NewEmailTemplateReadRepo(pool)

	return &BoundedContext{
		CreateTemplate: command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateTemplate: command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteTemplate: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetTemplate:    query.NewGetHandler(readRepo),
		ListTemplates:  query.NewListHandler(readRepo),
	}
}
