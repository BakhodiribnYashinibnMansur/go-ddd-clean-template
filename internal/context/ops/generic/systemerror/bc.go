package systemerror

import (
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"
	"gct/internal/context/ops/generic/systemerror/application/command"
	"gct/internal/context/ops/generic/systemerror/application/query"
	"gct/internal/context/ops/generic/systemerror/infrastructure/postgres"

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
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewSystemErrorWriteRepo(pool)
	readRepo := postgres.NewSystemErrorReadRepo(pool)

	return &BoundedContext{
		CreateSystemError: command.NewCreateSystemErrorHandler(writeRepo, committer, l),
		ResolveError:      command.NewResolveErrorHandler(writeRepo, committer, l),
		GetSystemError:    query.NewGetSystemErrorHandler(readRepo, l),
		ListSystemErrors:  query.NewListSystemErrorsHandler(readRepo, l),
	}
}
