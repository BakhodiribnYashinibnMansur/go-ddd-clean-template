package audit

import (
	"gct/internal/audit/application/command"
	"gct/internal/audit/application/query"
	"gct/internal/audit/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Audit BC.
type BoundedContext struct {
	// Commands
	CreateAuditLog        *command.CreateAuditLogHandler
	CreateEndpointHistory *command.CreateEndpointHistoryHandler

	// Queries
	ListAuditLogs       *query.ListAuditLogsHandler
	ListEndpointHistory *query.ListEndpointHistoryHandler
}

// NewBoundedContext creates a fully wired Audit bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	auditWriteRepo := postgres.NewAuditLogWriteRepo(pool)
	endpointWriteRepo := postgres.NewEndpointHistoryWriteRepo(pool)
	readRepo := postgres.NewAuditReadRepo(pool)

	return &BoundedContext{
		CreateAuditLog:        command.NewCreateAuditLogHandler(auditWriteRepo, eventBus, l),
		CreateEndpointHistory: command.NewCreateEndpointHistoryHandler(endpointWriteRepo, l),
		ListAuditLogs:         query.NewListAuditLogsHandler(readRepo),
		ListEndpointHistory:   query.NewListEndpointHistoryHandler(readRepo),
	}
}
