package query

import (
	"context"

	appdto "gct/internal/iprule/application"
	"gct/internal/iprule/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetIPRuleQuery holds the input for getting a single IP rule.
type GetIPRuleQuery struct {
	ID uuid.UUID
}

// GetIPRuleHandler handles the GetIPRuleQuery.
type GetIPRuleHandler struct {
	readRepo domain.IPRuleReadRepository
}

// NewGetIPRuleHandler creates a new GetIPRuleHandler.
func NewGetIPRuleHandler(readRepo domain.IPRuleReadRepository) *GetIPRuleHandler {
	return &GetIPRuleHandler{readRepo: readRepo}
}

// Handle executes the GetIPRuleQuery and returns an IPRuleView.
func (h *GetIPRuleHandler) Handle(ctx context.Context, q GetIPRuleQuery) (result *appdto.IPRuleView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetIPRuleHandler.Handle")
	defer func() { end(err) }()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.IPRuleView{
		ID:        v.ID,
		IPAddress: v.IPAddress,
		Action:    v.Action,
		Reason:    v.Reason,
		ExpiresAt: v.ExpiresAt,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
