package query

import (
	"context"

	appdto "gct/internal/notification/application"
	"gct/internal/notification/domain"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single notification.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.NotificationReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.NotificationReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns a NotificationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (*appdto.NotificationView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
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
