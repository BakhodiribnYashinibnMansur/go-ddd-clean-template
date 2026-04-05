package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/content/notification/application"
	"gct/internal/context/content/notification/domain"
	"gct/internal/platform/infrastructure/pgxutil"
)

// ListQuery holds the input for listing notifications with filtering.
type ListQuery struct {
	Filter domain.NotificationFilter
}

// ListResult holds the output of the list notifications query.
type ListResult struct {
	Notifications []*appdto.NotificationView
	Total         int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.NotificationReadRepository
	logger   logger.Log
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.NotificationReadRepository, l logger.Log) *ListHandler {
	return &ListHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListQuery and returns a list of NotificationView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (result *ListResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListNotifications", "notification")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "List", Entity: "notification", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	items := make([]*appdto.NotificationView, len(views))
	for i, v := range views {
		items[i] = &appdto.NotificationView{
			ID:        v.ID,
			UserID:    v.UserID,
			Title:     v.Title,
			Message:   v.Message,
			Type:      v.Type,
			ReadAt:    v.ReadAt,
			CreatedAt: v.CreatedAt,
		}
	}

	return &ListResult{
		Notifications: items,
		Total:         total,
	}, nil
}
