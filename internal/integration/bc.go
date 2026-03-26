package integration

import (
	appl "gct/internal/integration/application"
	"gct/internal/integration/application/command"
	"gct/internal/integration/application/query"
	"gct/internal/integration/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Integration BC.
type BoundedContext struct {
	// Commands
	CreateIntegration *command.CreateHandler
	UpdateIntegration *command.UpdateHandler
	DeleteIntegration *command.DeleteHandler

	// Queries
	GetIntegration   *query.GetHandler
	ListIntegrations *query.ListHandler

	// Services
	Cache *appl.CacheService
}

// NewBoundedContext creates a fully wired Integration bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewIntegrationWriteRepo(pool)
	readRepo := postgres.NewIntegrationReadRepo(pool)

	return &BoundedContext{
		CreateIntegration: command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateIntegration: command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteIntegration: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetIntegration:    query.NewGetHandler(readRepo),
		ListIntegrations:  query.NewListHandler(readRepo),
		Cache:             appl.NewCacheService(readRepo, l),
	}
}
