package query

import (
	"context"

	appdto "gct/internal/dataexport/application"
	"gct/internal/dataexport/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetDataExportQuery holds the input for getting a single data export.
type GetDataExportQuery struct {
	ID uuid.UUID
}

// GetDataExportHandler handles the GetDataExportQuery.
type GetDataExportHandler struct {
	readRepo domain.DataExportReadRepository
}

// NewGetDataExportHandler creates a new GetDataExportHandler.
func NewGetDataExportHandler(readRepo domain.DataExportReadRepository) *GetDataExportHandler {
	return &GetDataExportHandler{readRepo: readRepo}
}

// Handle executes the GetDataExportQuery and returns a DataExportView.
func (h *GetDataExportHandler) Handle(ctx context.Context, q GetDataExportQuery) (result *appdto.DataExportView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetDataExportHandler.Handle")
	defer func() { end(err) }()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
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
