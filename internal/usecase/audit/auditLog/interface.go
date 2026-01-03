package auditLog

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.AuditLog) error
	Gets(ctx context.Context, in *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error)
}
