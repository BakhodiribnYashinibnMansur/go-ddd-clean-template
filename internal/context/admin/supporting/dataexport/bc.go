package dataexport

import (
	"gct/internal/context/admin/supporting/dataexport/application/command"
	"gct/internal/context/admin/supporting/dataexport/application/query"
	"gct/internal/context/admin/supporting/dataexport/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the DataExport BC.
type BoundedContext struct {
	// Commands
	CreateDataExport *command.CreateDataExportHandler
	UpdateDataExport *command.UpdateDataExportHandler
	DeleteDataExport *command.DeleteDataExportHandler

	// Queries
	GetDataExport   *query.GetDataExportHandler
	ListDataExports *query.ListDataExportsHandler
}

// NewBoundedContext creates a fully wired DataExport bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewDataExportWriteRepo(pool)
	readRepo := postgres.NewDataExportReadRepo(pool)

	return &BoundedContext{
		CreateDataExport: command.NewCreateDataExportHandler(writeRepo, committer, l),
		UpdateDataExport: command.NewUpdateDataExportHandler(writeRepo, committer, l),
		DeleteDataExport: command.NewDeleteDataExportHandler(writeRepo, l),
		GetDataExport:    query.NewGetDataExportHandler(readRepo, l),
		ListDataExports:  query.NewListDataExportsHandler(readRepo, l),
	}
}
