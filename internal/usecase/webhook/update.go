package webhook

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	w, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.URL != nil {
		w.URL = *req.URL
	}
	if req.Secret != nil {
		w.Secret = *req.Secret
	}
	if req.Events != nil {
		w.Events = req.Events
	}
	if req.Headers != nil {
		w.Headers = req.Headers
	}
	if req.IsActive != nil {
		w.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}
