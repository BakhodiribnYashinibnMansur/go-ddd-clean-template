package client

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Approve godoc
// @Summary     Approve a pending user
// @Description Set user status to active (is_approved=true)
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID (UUID format)" format(uuid)
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid user ID"
// @Failure     401 {object} response.ErrorResponse "Unauthorized"
// @Failure     403 {object} response.ErrorResponse "Forbidden"
// @Failure     500 {object} response.ErrorResponse "Internal server error"
// @Security    BearerAuth
// @Router      /users/{user_id}/approve [post]
func (c *Controller) Approve(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - approve - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	if err := c.u.User.Client().Approve(ctx.Request.Context(), id.String()); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
