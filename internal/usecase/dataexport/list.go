package dataexport

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error) {
	return uc.repo.List(ctx, filter)
}
