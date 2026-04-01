package query

import (
	"context"

	appdto "gct/internal/integration/application"
	"gct/internal/integration/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single integration.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.IntegrationReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.IntegrationReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns an IntegrationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *appdto.IntegrationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.IntegrationView{
		ID:         view.ID,
		Name:       view.Name,
		Type:       view.Type,
		APIKey:     view.APIKey,
		WebhookURL: view.WebhookURL,
		Enabled:    view.Enabled,
		Config:     view.Config,
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
	}, nil
}
