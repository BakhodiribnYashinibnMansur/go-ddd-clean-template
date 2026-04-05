package audit

import (
	"gct/internal/context/iam/supporting/audit/application/command"
	"gct/internal/context/iam/supporting/audit/application/query"
	"gct/internal/context/iam/supporting/audit/application/subscriber"
	"gct/internal/context/iam/supporting/audit/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command, query, and subscriber handlers
// for the Audit BC.
type BoundedContext struct {
	// Commands
	CreateAuditLog        *command.CreateAuditLogHandler
	CreateEndpointHistory *command.CreateEndpointHistoryHandler

	// Queries
	ListAuditLogs       *query.ListAuditLogsHandler
	ListEndpointHistory *query.ListEndpointHistoryHandler

	// Subscribers (event-driven coupling with other BCs)
	userEvents *subscriber.UserEventSubscriber
}

// NewBoundedContext creates a fully wired Audit bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	auditWriteRepo := postgres.NewAuditLogWriteRepo(pool)
	endpointWriteRepo := postgres.NewEndpointHistoryWriteRepo(pool)
	readRepo := postgres.NewAuditReadRepo(pool)

	createAuditLog := command.NewCreateAuditLogHandler(auditWriteRepo, eventBus, l)

	return &BoundedContext{
		CreateAuditLog:        createAuditLog,
		CreateEndpointHistory: command.NewCreateEndpointHistoryHandler(endpointWriteRepo, l),
		ListAuditLogs:         query.NewListAuditLogsHandler(readRepo, l),
		ListEndpointHistory:   query.NewListEndpointHistoryHandler(readRepo, l),
		userEvents:            subscriber.NewUserEventSubscriber(createAuditLog, l),
	}
}

// RegisterSubscribers hooks this BC's event subscribers onto the shared event
// bus. Call from the composition root after the bus is ready and BEFORE the
// HTTP server starts serving traffic.
func (bc *BoundedContext) RegisterSubscribers(bus application.EventBus) error {
	return bc.userEvents.Register(bus)
}
