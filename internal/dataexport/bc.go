package dataexport

import (
	"gct/internal/dataexport/application/command"
	"gct/internal/dataexport/application/query"
	"gct/internal/dataexport/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

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
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewDataExportWriteRepo(pool)
	readRepo := postgres.NewDataExportReadRepo(pool)

	return &BoundedContext{
		CreateDataExport: command.NewCreateDataExportHandler(writeRepo, eventBus, l),
		UpdateDataExport: command.NewUpdateDataExportHandler(writeRepo, eventBus, l),
		DeleteDataExport: command.NewDeleteDataExportHandler(writeRepo, l),
		GetDataExport:    query.NewGetDataExportHandler(readRepo),
		ListDataExports:  query.NewListDataExportsHandler(readRepo),
	}
}
