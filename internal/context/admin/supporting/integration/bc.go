package integration

import (
	"time"

	appl "gct/internal/context/admin/supporting/integration/application"
	"gct/internal/context/admin/supporting/integration/application/command"
	"gct/internal/context/admin/supporting/integration/application/query"
	"gct/internal/context/admin/supporting/integration/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

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
	ValidateAPIKey   *query.ValidateAPIKeyHandler
	ResolveJWTAPIKey *query.ResolveJWTAPIKeyHandler

	// Repos — exposed for cross-cutting infrastructure (keyring bootstrap/rotation).
	ReadRepo  *postgres.IntegrationReadRepo
	WriteRepo *postgres.IntegrationWriteRepo

	// Services
	Cache *appl.CacheService
}

// NewBoundedContext creates a fully wired Integration bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, apiKeyPepper []byte, cacheTTL time.Duration, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewIntegrationWriteRepo(pool)
	readRepo := postgres.NewIntegrationReadRepo(pool)

	return &BoundedContext{
		CreateIntegration: command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateIntegration: command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteIntegration: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetIntegration:    query.NewGetHandler(readRepo, l),
		ListIntegrations:  query.NewListHandler(readRepo, l),
		ValidateAPIKey:    query.NewValidateAPIKeyHandler(readRepo, l),
		ResolveJWTAPIKey:  query.NewResolveJWTAPIKeyHandler(readRepo, apiKeyPepper, cacheTTL, l),
		ReadRepo:          readRepo,
		WriteRepo:         writeRepo,
		Cache:             appl.NewCacheService(readRepo, l),
	}
}
