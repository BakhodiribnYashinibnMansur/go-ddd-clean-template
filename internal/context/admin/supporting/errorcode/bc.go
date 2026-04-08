package errorcode

import (
	"gct/internal/context/admin/supporting/errorcode/application/command"
	"gct/internal/context/admin/supporting/errorcode/application/query"
	"gct/internal/context/admin/supporting/errorcode/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the ErrorCode BC.
type BoundedContext struct {
	// Commands
	CreateErrorCode *command.CreateErrorCodeHandler
	UpdateErrorCode *command.UpdateErrorCodeHandler
	DeleteErrorCode *command.DeleteErrorCodeHandler

	// Queries
	GetErrorCode   *query.GetErrorCodeHandler
	ListErrorCodes *query.ListErrorCodesHandler
}

// NewBoundedContext creates a fully wired ErrorCode bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewErrorCodeWriteRepo(pool)
	readRepo := postgres.NewErrorCodeReadRepo(pool)

	return &BoundedContext{
		CreateErrorCode: command.NewCreateErrorCodeHandler(writeRepo, committer, l),
		UpdateErrorCode: command.NewUpdateErrorCodeHandler(writeRepo, committer, l),
		DeleteErrorCode: command.NewDeleteErrorCodeHandler(writeRepo, committer, l),
		GetErrorCode:    query.NewGetErrorCodeHandler(readRepo, l),
		ListErrorCodes:  query.NewListErrorCodesHandler(readRepo, l),
	}
}
