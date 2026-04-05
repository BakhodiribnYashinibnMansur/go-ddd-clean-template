package query

import (
	"context"

	shared "gct/internal/platform/domain"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

// FindSessionQuery holds the input for fetching a session by ID.
type FindSessionQuery struct {
	SessionID uuid.UUID
}

// FindSessionHandler handles the FindSessionQuery.
type FindSessionHandler struct {
	readRepo domain.UserReadRepository
	logger   logger.Log
}

// NewFindSessionHandler creates a new FindSessionHandler.
func NewFindSessionHandler(readRepo domain.UserReadRepository, l logger.Log) *FindSessionHandler {
	return &FindSessionHandler{readRepo: readRepo, logger: l}
}

// Handle executes the FindSessionQuery and returns an AuthSession.
func (h *FindSessionHandler) Handle(ctx context.Context, q FindSessionQuery) (_ *shared.AuthSession, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "FindSessionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "FindSession", "user")()

	return h.readRepo.FindSessionByID(ctx, q.SessionID)
}
