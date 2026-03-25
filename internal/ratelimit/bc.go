package ratelimit

import (
	"gct/internal/ratelimit/application/command"
	"gct/internal/ratelimit/application/query"
	"gct/internal/ratelimit/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the RateLimit BC.
type BoundedContext struct {
	// Commands
	CreateRateLimit *command.CreateRateLimitHandler
	UpdateRateLimit *command.UpdateRateLimitHandler
	DeleteRateLimit *command.DeleteRateLimitHandler

	// Queries
	GetRateLimit   *query.GetRateLimitHandler
	ListRateLimits *query.ListRateLimitsHandler
}

// NewBoundedContext creates a fully wired RateLimit bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewRateLimitWriteRepo(pool)
	readRepo := postgres.NewRateLimitReadRepo(pool)

	return &BoundedContext{
		CreateRateLimit: command.NewCreateRateLimitHandler(writeRepo, eventBus, l),
		UpdateRateLimit: command.NewUpdateRateLimitHandler(writeRepo, eventBus, l),
		DeleteRateLimit: command.NewDeleteRateLimitHandler(writeRepo, l),
		GetRateLimit:    query.NewGetRateLimitHandler(readRepo),
		ListRateLimits:  query.NewListRateLimitsHandler(readRepo),
	}
}
