package emailtemplate

import (
	"context"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Test(ctx context.Context, id string, req domain.TestEmailTemplateRequest) (*domain.EmailLog, error) {
	tmpl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	templateID := tmpl.ID
	l := &domain.EmailLog{
		ID:         uuid.New().String(),
		TemplateID: &templateID,
		ToEmail:    req.ToEmail,
		Subject:    tmpl.Subject,
		Status:     "sent",
		SentAt:     &now,
	}
	if err := uc.repo.CreateLog(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}
