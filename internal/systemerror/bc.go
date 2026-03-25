package systemerror

import (
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/systemerror/application/command"
	"gct/internal/systemerror/application/query"
	"gct/internal/systemerror/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the SystemError BC.
type BoundedContext struct {
	// Commands
	CreateSystemError *command.CreateSystemErrorHandler
	ResolveError      *command.ResolveErrorHandler

	// Queries
	GetSystemError   *query.GetSystemErrorHandler
	ListSystemErrors *query.ListSystemErrorsHandler
}

// NewBoundedContext creates a fully wired SystemError bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewSystemErrorWriteRepo(pool)
	readRepo := postgres.NewSystemErrorReadRepo(pool)

	return &BoundedContext{
		CreateSystemError: command.NewCreateSystemErrorHandler(writeRepo, eventBus, l),
		ResolveError:      command.NewResolveErrorHandler(writeRepo, eventBus, l),
		GetSystemError:    query.NewGetSystemErrorHandler(readRepo),
		ListSystemErrors:  query.NewListSystemErrorsHandler(readRepo),
	}
}
