package translation

import (
	"gct/internal/context/content/generic/translation/application/command"
	"gct/internal/context/content/generic/translation/application/query"
	"gct/internal/context/content/generic/translation/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

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
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewTranslationWriteRepo(pool)
	readRepo := postgres.NewTranslationReadRepo(pool)

	return &BoundedContext{
		CreateTranslation: command.NewCreateTranslationHandler(writeRepo, committer, l),
		UpdateTranslation: command.NewUpdateTranslationHandler(writeRepo, committer, l),
		DeleteTranslation: command.NewDeleteTranslationHandler(writeRepo, l),
		GetTranslation:    query.NewGetTranslationHandler(readRepo, l),
		ListTranslations:  query.NewListTranslationsHandler(readRepo, l),
	}
}
