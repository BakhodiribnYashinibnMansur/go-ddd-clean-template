package query

import (
	"context"

	appdto "gct/internal/sitesetting/application"
	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

// GetSiteSettingQuery holds the input for getting a single site setting.
type GetSiteSettingQuery struct {
	ID uuid.UUID
}

// GetSiteSettingHandler handles the GetSiteSettingQuery.
type GetSiteSettingHandler struct {
	readRepo domain.SiteSettingReadRepository
}

// NewGetSiteSettingHandler creates a new GetSiteSettingHandler.
func NewGetSiteSettingHandler(readRepo domain.SiteSettingReadRepository) *GetSiteSettingHandler {
	return &GetSiteSettingHandler{readRepo: readRepo}
}

// Handle executes the GetSiteSettingQuery and returns a SiteSettingView.
func (h *GetSiteSettingHandler) Handle(ctx context.Context, q GetSiteSettingQuery) (*appdto.SiteSettingView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.SiteSettingView{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Type:        v.Type,
		Description: v.Description,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}, nil
}
