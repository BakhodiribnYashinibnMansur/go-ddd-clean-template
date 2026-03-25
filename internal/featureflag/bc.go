package featureflag

import (
	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the FeatureFlag BC.
type BoundedContext struct {
	// Commands
	CreateFlag *command.CreateHandler
	UpdateFlag *command.UpdateHandler
	DeleteFlag *command.DeleteHandler

	// Queries
	GetFlag   *query.GetHandler
	ListFlags *query.ListHandler
}

// NewBoundedContext creates a fully wired FeatureFlag bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewFeatureFlagWriteRepo(pool)
	readRepo := postgres.NewFeatureFlagReadRepo(pool)

	return &BoundedContext{
		CreateFlag: command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateFlag: command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteFlag: command.NewDeleteHandler(writeRepo, eventBus, l),
		GetFlag:    query.NewGetHandler(readRepo),
		ListFlags:  query.NewListHandler(readRepo),
	}
}
