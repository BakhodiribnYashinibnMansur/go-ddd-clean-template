package query

import (
	"context"

	appdto "gct/internal/announcement/application"
	"gct/internal/announcement/domain"

	"github.com/google/uuid"
)

// GetAnnouncementQuery holds the input for getting a single announcement.
type GetAnnouncementQuery struct {
	ID uuid.UUID
}

// GetAnnouncementHandler handles the GetAnnouncementQuery.
type GetAnnouncementHandler struct {
	readRepo domain.AnnouncementReadRepository
}

// NewGetAnnouncementHandler creates a new GetAnnouncementHandler.
func NewGetAnnouncementHandler(readRepo domain.AnnouncementReadRepository) *GetAnnouncementHandler {
	return &GetAnnouncementHandler{readRepo: readRepo}
}

// Handle executes the GetAnnouncementQuery and returns an AnnouncementView.
func (h *GetAnnouncementHandler) Handle(ctx context.Context, q GetAnnouncementQuery) (*appdto.AnnouncementView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return toAppView(v), nil
}
