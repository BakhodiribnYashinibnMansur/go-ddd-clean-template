package auditlog

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.AuditLog) error
	Gets(ctx context.Context, in *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error)
	GetLogins(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.LoginEntry, int, error)
	GetSessions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.SessionEntry, int, error)
	GetActions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.ActionEntry, int, error)
}
