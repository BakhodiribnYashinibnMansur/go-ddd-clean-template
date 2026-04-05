package file

import (
	"gct/internal/context/content/file/application/command"
	"gct/internal/context/content/file/application/query"
	"gct/internal/context/content/file/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

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
		GetFile:    query.NewGetFileHandler(readRepo, l),
		ListFiles:  query.NewListFilesHandler(readRepo, l),
	}
}
