package query

import (
	"context"

	appdto "gct/internal/webhook/application"
	"gct/internal/webhook/domain"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single webhook.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.WebhookReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.WebhookReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns a WebhookView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (*appdto.WebhookView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.WebhookView{
		ID:        view.ID,
		Name:      view.Name,
		URL:       view.URL,
		Secret:    view.Secret,
		Events:    view.Events,
		Enabled:   view.Enabled,
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}, nil
}
