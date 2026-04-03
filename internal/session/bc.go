package session

import (
	"gct/internal/session/application/command"
	"gct/internal/session/application/query"
	"gct/internal/session/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all handlers for the Session BC.
type BoundedContext struct {
	// Commands
	RevokeSession     *command.RevokeSessionHandler
	RevokeAllSessions *command.RevokeAllSessionsHandler

	// Queries
	GetSession   *query.GetSessionHandler
	ListSessions *query.ListSessionsHandler
}

// NewBoundedContext creates a fully wired Session bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	readRepo := postgres.NewSessionReadRepo(pool)

	return &BoundedContext{
		RevokeSession:     command.NewRevokeSessionHandler(eventBus, l),
		RevokeAllSessions: command.NewRevokeAllSessionsHandler(eventBus, l),
		GetSession:        query.NewGetSessionHandler(readRepo, l),
		ListSessions:      query.NewListSessionsHandler(readRepo, l),
	}
}
