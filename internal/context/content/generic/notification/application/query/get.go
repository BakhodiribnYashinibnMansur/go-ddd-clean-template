package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/generic/notification/application/dto"
	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	notifrepo "gct/internal/context/content/generic/notification/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single notification.
type GetQuery struct {
	ID notifentity.NotificationID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo notifrepo.NotificationReadRepository
	logger   logger.Log
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo notifrepo.NotificationReadRepository, l logger.Log) *GetHandler {
	return &GetHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetQuery and returns a NotificationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *dto.NotificationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetNotification", "notification")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "Get", Entity: "notification", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.NotificationView{
		ID:        uuid.UUID(view.ID),
		UserID:    view.UserID,
		Title:     view.Title,
		Message:   view.Message,
		Type:      view.Type,
		ReadAt:    view.ReadAt,
		CreatedAt: view.CreatedAt,
	}, nil
}
