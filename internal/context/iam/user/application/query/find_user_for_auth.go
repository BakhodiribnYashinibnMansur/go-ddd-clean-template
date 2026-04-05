package query

import (
	"context"

	shared "gct/internal/platform/domain"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

// FindUserForAuthQuery holds the input for fetching minimal user data for auth.
type FindUserForAuthQuery struct {
	UserID uuid.UUID
}

// FindUserForAuthHandler handles the FindUserForAuthQuery.
type FindUserForAuthHandler struct {
	readRepo domain.UserReadRepository
	logger   logger.Log
}

// NewFindUserForAuthHandler creates a new FindUserForAuthHandler.
func NewFindUserForAuthHandler(readRepo domain.UserReadRepository, l logger.Log) *FindUserForAuthHandler {
	return &FindUserForAuthHandler{readRepo: readRepo, logger: l}
}

// Handle executes the FindUserForAuthQuery and returns an AuthUser.
func (h *FindUserForAuthHandler) Handle(ctx context.Context, q FindUserForAuthQuery) (_ *shared.AuthUser, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "FindUserForAuthHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "FindUserForAuth", "user")()

	return h.readRepo.FindUserForAuth(ctx, q.UserID)
}
