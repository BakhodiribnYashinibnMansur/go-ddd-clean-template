package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/admin/dataexport/application"
	"gct/internal/context/admin/dataexport/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetDataExportQuery holds the input for getting a single data export.
type GetDataExportQuery struct {
	ID domain.DataExportID
}

// GetDataExportHandler handles the GetDataExportQuery.
type GetDataExportHandler struct {
	readRepo domain.DataExportReadRepository
	logger   logger.Log
}

// NewGetDataExportHandler creates a new GetDataExportHandler.
func NewGetDataExportHandler(readRepo domain.DataExportReadRepository, l logger.Log) *GetDataExportHandler {
	return &GetDataExportHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetDataExportQuery and returns a DataExportView.
func (h *GetDataExportHandler) Handle(ctx context.Context, q GetDataExportQuery) (result *appdto.DataExportView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetDataExportHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetDataExport", "data_export")()

	v, err := h.readRepo.FindByID(ctx, q.ID.UUID())
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetDataExport", Entity: "data_export", EntityID: q.ID.UUID(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.DataExportView{
		ID:        v.ID,
		UserID:    v.UserID,
		DataType:  v.DataType,
		Format:    v.Format,
		Status:    v.Status,
		FileURL:   v.FileURL,
		Error:     v.Error,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
