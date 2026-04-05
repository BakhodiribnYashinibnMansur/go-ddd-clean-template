package query

import (
	"context"
	"fmt"

	"gct/internal/context/iam/user/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// FindSessionQuery holds the input for fetching a session by ID.
type FindSessionQuery struct {
	SessionID domain.SessionID
}

// FindSessionHandler handles the FindSessionQuery.
type FindSessionHandler struct {
	readRepo domain.UserReadRepository
	logger   queryLogger
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

	session, err := h.readRepo.FindSessionByID(ctx, q.SessionID.UUID())
	if err != nil {
		return nil, fmt.Errorf("find_session: read repo: %w", err)
	}
	return session, nil
}
