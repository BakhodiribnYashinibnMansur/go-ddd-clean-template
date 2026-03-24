package auditlog

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"go.uber.org/zap"
)

func (uc *UseCase) Gets(ctx context.Context, in *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error) {
	logs, count, err := uc.repo.Postgres.Audit.Log.Gets(ctx, in)
	if err != nil {
		uc.logger.Errorw("audit log retrieval failed", zap.Error(err))
		return nil, 0, apperrors.MapRepoToServiceError(err)
	}

	return logs, count, nil
}
