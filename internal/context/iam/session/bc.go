package session

import (
	"gct/internal/context/iam/session/application/query"
	"gct/internal/context/iam/session/infrastructure/postgres"
	"gct/internal/platform/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all query handlers for the Session read-only BC.
type BoundedContext struct {
	// Queries
	GetSession   *query.GetSessionHandler
	ListSessions *query.ListSessionsHandler
}

// NewBoundedContext creates a fully wired Session bounded context (read-only).
func NewBoundedContext(pool *pgxpool.Pool, l logger.Log) *BoundedContext {
	readRepo := postgres.NewSessionReadRepo(pool)

	return &BoundedContext{
		GetSession:   query.NewGetSessionHandler(readRepo, l),
		ListSessions: query.NewListSessionsHandler(readRepo, l),
	}
}
