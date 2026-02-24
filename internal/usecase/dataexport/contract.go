package dataexport

import (
	"context"

	"gct/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, e *domain.DataExport) error
	List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error)
	Delete(ctx context.Context, id string) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateDataExportRequest, userID string) (*domain.DataExport, error)
	List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error)
	Delete(ctx context.Context, id string) error
}
