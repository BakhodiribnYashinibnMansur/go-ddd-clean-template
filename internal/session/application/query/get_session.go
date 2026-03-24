package query

import (
	"context"

	appdto "gct/internal/session/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// GetSessionQuery holds the input for fetching a single session.
type GetSessionQuery struct {
	ID uuid.UUID
}

// GetSessionHandler handles the GetSessionQuery.
type GetSessionHandler struct {
	repo SessionReadRepository
	l    logger.Log
}

// NewGetSessionHandler creates a new GetSessionHandler.
func NewGetSessionHandler(repo SessionReadRepository, l logger.Log) *GetSessionHandler {
	return &GetSessionHandler{repo: repo, l: l}
}

// Handle executes the GetSessionQuery and returns a SessionView.
func (h *GetSessionHandler) Handle(ctx context.Context, q GetSessionQuery) (*appdto.SessionView, error) {
	view, err := h.repo.FindByID(ctx, q.ID)
	if err != nil {
		h.l.Errorc(ctx, "session.query.GetSession failed", "session_id", q.ID, "error", err)
		return nil, err
	}
	return view, nil
}
