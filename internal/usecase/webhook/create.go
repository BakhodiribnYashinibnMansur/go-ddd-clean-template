package webhook

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateWebhookRequest) (*domain.Webhook, error) {
	w := &domain.Webhook{
		ID:       uuid.New(),
		Name:     req.Name,
		URL:      req.URL,
		Secret:   req.Secret,
		Events:   req.Events,
		Headers:  req.Headers,
		IsActive: req.IsActive,
	}
	if w.Events == nil {
		w.Events = []string{}
	}
	if w.Headers == nil {
		w.Headers = map[string]any{}
	}
	if err := uc.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}
