package query

import (
	"context"

	appdto "gct/internal/emailtemplate/application"
	"gct/internal/emailtemplate/domain"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single email template.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.EmailTemplateReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.EmailTemplateReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns an EmailTemplateView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (*appdto.EmailTemplateView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.EmailTemplateView{
		ID:        view.ID,
		Name:      view.Name,
		Subject:   view.Subject,
		HTMLBody:  view.HTMLBody,
		TextBody:  view.TextBody,
		Variables: view.Variables,
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}, nil
}
