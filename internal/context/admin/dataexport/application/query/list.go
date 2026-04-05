package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/admin/dataexport/application"
	"gct/internal/context/admin/dataexport/domain"
	"gct/internal/platform/infrastructure/pgxutil"
)

// ListDataExportsQuery holds the input for listing data exports with filtering.
type ListDataExportsQuery struct {
	Filter domain.DataExportFilter
}

// ListDataExportsResult holds the output of the list data exports query.
type ListDataExportsResult struct {
	Exports []*appdto.DataExportView
	Total   int64
}

// ListDataExportsHandler handles the ListDataExportsQuery.
type ListDataExportsHandler struct {
	readRepo domain.DataExportReadRepository
	logger   logger.Log
}

// NewListDataExportsHandler creates a new ListDataExportsHandler.
func NewListDataExportsHandler(readRepo domain.DataExportReadRepository, l logger.Log) *ListDataExportsHandler {
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

	result := make([]*appdto.DataExportView, len(views))
	for i, v := range views {
		result[i] = &appdto.DataExportView{
			ID:        v.ID,
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
