package emailtemplate

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) Update(ctx context.Context, id string, req domain.UpdateEmailTemplateRequest) (*domain.EmailTemplate, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Subject != nil {
		t.Subject = *req.Subject
	}
	if req.HtmlBody != nil {
		t.HtmlBody = *req.HtmlBody
	}
	if req.TextBody != nil {
		t.TextBody = *req.TextBody
	}
	if req.Type != nil {
		t.Type = *req.Type
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
