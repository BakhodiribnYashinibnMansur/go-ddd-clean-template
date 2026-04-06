package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/dataexport/application/dto"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListDataExportsQuery holds the input for listing data exports with filtering.
type ListDataExportsQuery struct {
	Filter exportrepo.DataExportFilter
}

// ListDataExportsResult holds the output of the list data exports query.
type ListDataExportsResult struct {
	Exports []*dto.DataExportView
	Total   int64
}

// ListDataExportsHandler handles the ListDataExportsQuery.
type ListDataExportsHandler struct {
	readRepo exportrepo.DataExportReadRepository
	logger   logger.Log
}

// NewListDataExportsHandler creates a new ListDataExportsHandler.
func NewListDataExportsHandler(readRepo exportrepo.DataExportReadRepository, l logger.Log) *ListDataExportsHandler {
	return &ListDataExportsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListDataExportsQuery and returns a list of DataExportView with total count.
func (h *ListDataExportsHandler) Handle(ctx context.Context, q ListDataExportsQuery) (_ *ListDataExportsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListDataExportsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListDataExports", "data_export")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListDataExports", Entity: "data_export", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.DataExportView, len(views))
	for i, v := range views {
		result[i] = &dto.DataExportView{
			ID:        uuid.UUID(v.ID),
			UserID:    v.UserID,
			DataType:  v.DataType,
			Format:    v.Format,
			Status:    v.Status,
			FileURL:   v.FileURL,
			Error:     v.Error,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListDataExportsResult{
		Exports: result,
		Total:   total,
	}, nil
}
