package query

import (
	"context"

	shared "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// FindUserForAuthQuery holds the input for fetching minimal user data for auth.
type FindUserForAuthQuery struct {
	UserID uuid.UUID
}

// FindUserForAuthHandler handles the FindUserForAuthQuery.
type FindUserForAuthHandler struct {
	readRepo domain.UserReadRepository
}

// NewFindUserForAuthHandler creates a new FindUserForAuthHandler.
func NewFindUserForAuthHandler(readRepo domain.UserReadRepository) *FindUserForAuthHandler {
	return &FindUserForAuthHandler{readRepo: readRepo}
}

// Handle executes the FindUserForAuthQuery and returns an AuthUser.
func (h *FindUserForAuthHandler) Handle(ctx context.Context, q FindUserForAuthQuery) (_ *shared.AuthUser, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "FindUserForAuthHandler.Handle")
	defer func() { end(err) }()

	return h.readRepo.FindUserForAuth(ctx, q.UserID)
}
