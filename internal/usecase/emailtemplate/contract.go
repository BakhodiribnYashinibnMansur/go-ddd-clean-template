package emailtemplate

import (
	"context"

	"gct/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, t *domain.EmailTemplate) error
	GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error)
	List(ctx context.Context, filter domain.EmailTemplateFilter) ([]domain.EmailTemplate, int64, error)
	Update(ctx context.Context, t *domain.EmailTemplate) error
	Delete(ctx context.Context, id string) error
	CreateLog(ctx context.Context, l *domain.EmailLog) error
	ListLogs(ctx context.Context, filter domain.EmailLogFilter) ([]domain.EmailLog, int64, error)
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateEmailTemplateRequest) (*domain.EmailTemplate, error)
	GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error)
	List(ctx context.Context, filter domain.EmailTemplateFilter) ([]domain.EmailTemplate, int64, error)
	Update(ctx context.Context, id string, req domain.UpdateEmailTemplateRequest) (*domain.EmailTemplate, error)
	Delete(ctx context.Context, id string) error
	Test(ctx context.Context, id string, req domain.TestEmailTemplateRequest) (*domain.EmailLog, error)
	ListLogs(ctx context.Context, filter domain.EmailLogFilter) ([]domain.EmailLog, int64, error)
}
