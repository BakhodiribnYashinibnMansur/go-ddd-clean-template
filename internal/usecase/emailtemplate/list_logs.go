package emailtemplate

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) ListLogs(ctx context.Context, filter domain.EmailLogFilter) ([]domain.EmailLog, int64, error) {
	return uc.repo.ListLogs(ctx, filter)
}
