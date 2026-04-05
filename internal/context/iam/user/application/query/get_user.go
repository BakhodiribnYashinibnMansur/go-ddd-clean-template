package query

import (
	"context"

	appdto "gct/internal/context/iam/user/application"
	"gct/internal/context/iam/user/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetUserQuery holds the input for fetching a single user.
type GetUserQuery struct {
	ID domain.UserID
}

// GetUserHandler handles the GetUserQuery.
type GetUserHandler struct {
	readRepo domain.UserReadRepository
	logger   queryLogger
}

// NewGetUserHandler creates a new GetUserHandler.
func NewGetUserHandler(readRepo domain.UserReadRepository, l logger.Log) *GetUserHandler {
	return &GetUserHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetUserQuery and returns a UserView.
func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (_ *appdto.UserView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetUserHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetUser", "user")()

	view, err := h.readRepo.FindByID(ctx, q.ID.UUID())
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetUser", Entity: "user", EntityID: q.ID.UUID(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	// Map domain UserView to application UserView.
	return &appdto.UserView{
		ID:         view.ID,
		Phone:      view.Phone,
		Email:      view.Email,
		Username:   view.Username,
		RoleID:     view.RoleID,
		Attributes: view.Attributes,
		Active:     view.Active,
		IsApproved: view.IsApproved,
	}, nil
}
