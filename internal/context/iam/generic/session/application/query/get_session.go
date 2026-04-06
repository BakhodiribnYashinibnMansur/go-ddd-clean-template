package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"

	"gct/internal/context/iam/generic/session/application/dto"
	"gct/internal/context/iam/generic/session/domain/entity"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetSessionQuery holds the input for fetching a single session.
type GetSessionQuery struct {
	ID entity.SessionID
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
func (h *GetSessionHandler) Handle(ctx context.Context, q GetSessionQuery) (_ *dto.SessionView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSessionHandler.Handle")
	defer func() { end(err) }()

	view, err := h.repo.FindByID(ctx, q.ID.UUID())
	if err != nil {
		h.l.Errorc(ctx, "session.query.GetSession failed", "session_id", q.ID, "error", err)
		return nil, apperrors.MapToServiceError(err)
	}
	return view, nil
}
