package query

import (
	"context"
	"fmt"

	"gct/internal/context/iam/generic/user/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// FindUserForAuthQuery holds the input for fetching minimal user data for auth.
type FindUserForAuthQuery struct {
	UserID domain.UserID
}

// FindUserForAuthHandler handles the FindUserForAuthQuery.
type FindUserForAuthHandler struct {
	readRepo domain.UserReadRepository
	logger   queryLogger
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

	user, err := h.readRepo.FindUserForAuth(ctx, q.UserID.UUID())
	if err != nil {
		return nil, fmt.Errorf("find_user_for_auth: read repo: %w", err)
	}
	return user, nil
}
