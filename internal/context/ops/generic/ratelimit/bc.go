package ratelimit

import (
	"gct/internal/context/ops/generic/ratelimit/application/command"
	"gct/internal/context/ops/generic/ratelimit/application/query"
	"gct/internal/context/ops/generic/ratelimit/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

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
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewRateLimitWriteRepo(pool)
	readRepo := postgres.NewRateLimitReadRepo(pool)

	return &BoundedContext{
		CreateRateLimit: command.NewCreateRateLimitHandler(writeRepo, committer, l),
		UpdateRateLimit: command.NewUpdateRateLimitHandler(writeRepo, committer, l),
		DeleteRateLimit: command.NewDeleteRateLimitHandler(writeRepo, l),
		GetRateLimit:    query.NewGetRateLimitHandler(readRepo, l),
		ListRateLimits:  query.NewListRateLimitsHandler(readRepo, l),
	}
}
