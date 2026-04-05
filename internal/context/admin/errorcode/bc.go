package errorcode

import (
	"gct/internal/context/admin/errorcode/application/command"
	"gct/internal/context/admin/errorcode/application/query"
	"gct/internal/context/admin/errorcode/infrastructure/postgres"
	"gct/internal/platform/application"
	"gct/internal/platform/infrastructure/logger"

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
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewErrorCodeWriteRepo(pool)
	readRepo := postgres.NewErrorCodeReadRepo(pool)

	return &BoundedContext{
		CreateErrorCode: command.NewCreateErrorCodeHandler(writeRepo, eventBus, l),
		UpdateErrorCode: command.NewUpdateErrorCodeHandler(writeRepo, eventBus, l),
		DeleteErrorCode: command.NewDeleteErrorCodeHandler(writeRepo, eventBus, l),
		GetErrorCode:    query.NewGetErrorCodeHandler(readRepo, l),
		ListErrorCodes:  query.NewListErrorCodesHandler(readRepo, l),
	}
}
