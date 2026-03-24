package client

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete user
// @Description Delete user by ID (soft delete)
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /users/{user_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - delete - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "User deleted successfully") {
		return
	}

	err = c.u.User.Client().Delete(ctx.Request.Context(), &domain.UserFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
