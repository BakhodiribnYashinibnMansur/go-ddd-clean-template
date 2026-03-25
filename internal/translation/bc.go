package translation

import (
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/translation/application/command"
	"gct/internal/translation/application/query"
	"gct/internal/translation/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Translation BC.
type BoundedContext struct {
	// Commands
	CreateTranslation *command.CreateTranslationHandler
	UpdateTranslation *command.UpdateTranslationHandler
	DeleteTranslation *command.DeleteTranslationHandler

	// Queries
	GetTranslation   *query.GetTranslationHandler
	ListTranslations *query.ListTranslationsHandler
}

// NewBoundedContext creates a fully wired Translation bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewTranslationWriteRepo(pool)
	readRepo := postgres.NewTranslationReadRepo(pool)

	return &BoundedContext{
		CreateTranslation: command.NewCreateTranslationHandler(writeRepo, eventBus, l),
		UpdateTranslation: command.NewUpdateTranslationHandler(writeRepo, eventBus, l),
		DeleteTranslation: command.NewDeleteTranslationHandler(writeRepo, l),
		GetTranslation:    query.NewGetTranslationHandler(readRepo),
		ListTranslations:  query.NewListTranslationsHandler(readRepo),
	}
}
