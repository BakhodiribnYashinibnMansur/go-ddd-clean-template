package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// ChangeRole godoc
// @Summary     Change user role
// @Description Update the role assigned to a user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID (UUID format)" format(uuid)
// @Param       body body domain.ChangeRoleRequest true "Role change request"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid request"
// @Failure     401 {object} response.ErrorResponse "Unauthorized"
// @Failure     403 {object} response.ErrorResponse "Forbidden"
// @Failure     500 {object} response.ErrorResponse "Internal server error"
// @Security    BearerAuth
// @Router      /users/{user_id}/role [post]
func (c *Controller) ChangeRole(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - change_role - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	var req domain.ChangeRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	if err := c.u.User.Client().ChangeRole(ctx.Request.Context(), id.String(), req.Role); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
