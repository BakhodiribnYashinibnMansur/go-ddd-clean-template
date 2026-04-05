package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/content/notification/application"
	"gct/internal/context/content/notification/domain"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single notification.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.NotificationReadRepository
	logger   logger.Log
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.NotificationReadRepository, l logger.Log) *GetHandler {
	return &GetHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetQuery and returns a NotificationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *appdto.NotificationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetNotification", "notification")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "Get", Entity: "notification", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.NotificationView{
		ID:        view.ID,
		UserID:    view.UserID,
		Title:     view.Title,
		Message:   view.Message,
		Type:      view.Type,
		ReadAt:    view.ReadAt,
		CreatedAt: view.CreatedAt,
	}, nil
}
