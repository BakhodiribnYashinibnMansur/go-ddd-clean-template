package query

import (
	"context"

	appdto "gct/internal/session/application"
	"gct/internal/shared/infrastructure/logger"
)

// ListSessionsQuery holds the input for listing sessions with filtering.
type ListSessionsQuery struct {
	Filter appdto.SessionsFilter
}

// ListSessionsResult holds the output of the list sessions query.
type ListSessionsResult struct {
	Sessions []*appdto.SessionView
	Total    int64
}

// ListSessionsHandler handles the ListSessionsQuery.
type ListSessionsHandler struct {
	repo SessionReadRepository
	l    logger.Log
}

// NewListSessionsHandler creates a new ListSessionsHandler.
func NewListSessionsHandler(repo SessionReadRepository, l logger.Log) *ListSessionsHandler {
	return &ListSessionsHandler{repo: repo, l: l}
}

// Handle executes the ListSessionsQuery and returns a list of SessionView with total count.
func (h *ListSessionsHandler) Handle(ctx context.Context, q ListSessionsQuery) (*ListSessionsResult, error) {
	views, total, err := h.repo.List(ctx, q.Filter)
	if err != nil {
		h.l.Errorc(ctx, "session.query.ListSessions failed", "error", err)
		return nil, err
	}

	return &ListSessionsResult{
		Sessions: views,
		Total:    total,
	}, nil
}
