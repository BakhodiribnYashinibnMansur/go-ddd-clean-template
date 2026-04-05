package session

import (
	"gct/internal/context/iam/generic/session/application/command"
	"gct/internal/context/iam/generic/session/application/query"
	"gct/internal/context/iam/generic/session/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

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
