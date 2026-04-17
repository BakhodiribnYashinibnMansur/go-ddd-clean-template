package repository

import (
	"context"

	"gct/internal/context/admin/supporting/dataexport/domain/entity"
	shareddomain "gct/internal/kernel/domain"
)

// DataExportRepository is the write-side repository for the DataExport aggregate.
// Implementations must return ErrDataExportNotFound from FindByID when no row matches.
// Update must persist status transitions and file URL changes atomically.
type DataExportRepository interface {
	Save(ctx context.Context, q shareddomain.Querier, entity *entity.DataExport) error
	Update(ctx context.Context, q shareddomain.Querier, entity *entity.DataExport) error
	FindByID(ctx context.Context, id entity.DataExportID) (*entity.DataExport, error)
	Delete(ctx context.Context, q shareddomain.Querier, id entity.DataExportID) error
}
