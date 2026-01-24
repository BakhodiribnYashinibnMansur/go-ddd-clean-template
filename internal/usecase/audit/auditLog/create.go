package auditLog

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"go.uber.org/zap"
)

func (uc *UseCase) Create(ctx context.Context, in *domain.AuditLog) error {
	err := uc.repo.Postgres.Audit.Log.Create(ctx, in)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("audit log creation failed", zap.Error(err))
		return apperrors.MapRepoToServiceError(ctx, err)
	}
	return nil
}
