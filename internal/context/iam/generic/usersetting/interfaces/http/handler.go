package http

import (
	"net/http"

	"gct/internal/context/iam/generic/usersetting"
	"gct/internal/context/iam/generic/usersetting/application/command"
	"gct/internal/context/iam/generic/usersetting/application/query"
	"gct/internal/context/iam/generic/usersetting/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the UserSetting bounded context.
type Handler struct {
	bc *usersetting.BoundedContext
	l  logger.Log
}

// NewHandler creates a new UserSetting HTTP handler.
func NewHandler(bc *usersetting.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Upsert creates or updates a user setting.
func (h *Handler) Upsert(ctx *gin.Context) {
	var req UpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpsertUserSettingCommand{
		UserID: req.UserID,
		Key:    req.Key,
		Value:  req.Value,
	}
	if err := h.bc.UpsertUserSetting.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of user settings.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListUserSettingsQuery{
		Filter: domain.UserSettingFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListUserSettings.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Settings, "total": result.Total})
}

// Delete deletes a user setting.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseUserSettingID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteUserSetting.Handle(ctx.Request.Context(), command.DeleteUserSettingCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
