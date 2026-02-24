package emailtemplate

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateEmailTemplateRequest) (*domain.EmailTemplate, error) {
	t := &domain.EmailTemplate{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Subject:  req.Subject,
		HtmlBody: req.HtmlBody,
		TextBody: req.TextBody,
		Type:     req.Type,
		IsActive: true,
	}
	if err := uc.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
