package file

import (
	"gct/internal/file/application/command"
	"gct/internal/file/application/query"
	"gct/internal/file/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the File BC.
type BoundedContext struct {
	// Commands
	CreateFile *command.CreateFileHandler

	// Queries
	GetFile   *query.GetFileHandler
	ListFiles *query.ListFilesHandler
}

// NewBoundedContext creates a fully wired File bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewFileWriteRepo(pool)
	readRepo := postgres.NewFileReadRepo(pool)

	return &BoundedContext{
		CreateFile: command.NewCreateFileHandler(writeRepo, eventBus, l),
		GetFile:    query.NewGetFileHandler(readRepo),
		ListFiles:  query.NewListFilesHandler(readRepo),
	}
}
