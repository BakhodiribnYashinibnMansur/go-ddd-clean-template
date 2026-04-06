package query

import (
	"context"
	"fmt"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// FindSessionQuery holds the input for fetching a session by ID.
type FindSessionQuery struct {
	SessionID userentity.SessionID
}

// FindSessionHandler handles the FindSessionQuery.
type FindSessionHandler struct {
	readRepo userrepo.UserReadRepository
	logger   queryLogger
}

// NewFindSessionHandler creates a new FindSessionHandler.
func NewFindSessionHandler(readRepo userrepo.UserReadRepository, l logger.Log) *FindSessionHandler {
	return &FindSessionHandler{readRepo: readRepo, logger: l}
}

// Handle executes the FindSessionQuery and returns an AuthSession.
func (h *FindSessionHandler) Handle(ctx context.Context, q FindSessionQuery) (_ *shared.AuthSession, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "FindSessionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "FindSession", "user")()

	session, err := h.readRepo.FindSessionByID(ctx, q.SessionID)
	if err != nil {
		return nil, fmt.Errorf("find_session: read repo: %w", err)
	}
	return session, nil
}
