package query

import (
	"context"

	shared "gct/internal/shared/domain"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// FindSessionQuery holds the input for fetching a session by ID.
type FindSessionQuery struct {
	SessionID uuid.UUID
}

// FindSessionHandler handles the FindSessionQuery.
type FindSessionHandler struct {
	readRepo domain.UserReadRepository
}

// NewFindSessionHandler creates a new FindSessionHandler.
func NewFindSessionHandler(readRepo domain.UserReadRepository) *FindSessionHandler {
	return &FindSessionHandler{readRepo: readRepo}
}

// Handle executes the FindSessionQuery and returns an AuthSession.
func (h *FindSessionHandler) Handle(ctx context.Context, q FindSessionQuery) (*shared.AuthSession, error) {
	return h.readRepo.FindSessionByID(ctx, q.SessionID)
}
