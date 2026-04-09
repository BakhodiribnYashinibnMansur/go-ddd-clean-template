package activitylog

import (
	"gct/internal/context/ops/supporting/activitylog/application/command"
	"gct/internal/context/ops/supporting/activitylog/application/query"
	"gct/internal/context/ops/supporting/activitylog/application/subscriber"
	"gct/internal/context/ops/supporting/activitylog/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command, query, and subscriber handlers
// for the ActivityLog BC.
type BoundedContext struct {
	// Queries
	ListActivityLogs *query.ListActivityLogsHandler

	// Subscribers (event-driven coupling with other BCs)
	userEvents *subscriber.UserEventSubscriber
	generic    *subscriber.GenericSubscriber
}

// NewBoundedContext creates a fully wired ActivityLog bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewActivityLogWriteRepo(pool)
	readRepo := postgres.NewActivityLogReadRepo(pool)

	createBatch := command.NewCreateActivityLogBatchHandler(writeRepo, l)

	return &BoundedContext{
		ListActivityLogs: query.NewListActivityLogsHandler(readRepo, l),
		userEvents:       subscriber.NewUserEventSubscriber(createBatch, l),
		generic:          subscriber.NewGenericSubscriber(createBatch, l),
	}
}

// RegisterSubscribers hooks this BC's event subscribers onto the shared event bus.
func (bc *BoundedContext) RegisterSubscribers(bus application.EventBus) error {
	if err := bc.userEvents.Register(bus); err != nil {
		return err
	}
	return bc.generic.Register(bus)
}
