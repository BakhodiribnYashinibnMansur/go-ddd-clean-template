package query

import (
	"context"

	appdto "gct/internal/file/application"
	"gct/internal/file/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetFileQuery holds the input for getting a single file.
type GetFileQuery struct {
	ID uuid.UUID
}

// GetFileHandler handles the GetFileQuery.
type GetFileHandler struct {
	readRepo domain.FileReadRepository
}

// NewGetFileHandler creates a new GetFileHandler.
func NewGetFileHandler(readRepo domain.FileReadRepository) *GetFileHandler {
	return &GetFileHandler{readRepo: readRepo}
}

// Handle executes the GetFileQuery and returns a FileView.
func (h *GetFileHandler) Handle(ctx context.Context, q GetFileQuery) (result *appdto.FileView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetFileHandler.Handle")
	defer func() { end(err) }()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.FileView{
		ID:           v.ID,
		Name:         v.Name,
		OriginalName: v.OriginalName,
		MimeType:     v.MimeType,
		Size:         v.Size,
		Path:         v.Path,
		URL:          v.URL,
		UploadedBy:   v.UploadedBy,
		CreatedAt:    v.CreatedAt,
	}, nil
}
