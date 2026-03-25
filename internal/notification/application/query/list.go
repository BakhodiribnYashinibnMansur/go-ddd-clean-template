package query

import (
	"context"

	appdto "gct/internal/notification/application"
	"gct/internal/notification/domain"
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
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.NotificationReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of NotificationView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.NotificationView, len(views))
	for i, v := range views {
		result[i] = &appdto.NotificationView{
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
		Notifications: result,
		Total:         total,
	}, nil
}
