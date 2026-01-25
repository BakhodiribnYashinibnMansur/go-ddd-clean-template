package role

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"github.com/gin-gonic/gin"
)

// Assign godoc
// @Summary     Assign role to user
// @Description Assign a role to a user
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID"
// @Param       role_id path string true "Role ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/users/{user_id}/roles/{role_id} [post]
func (c *Controller) Assign(ctx *gin.Context) {
	userID, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - assign - user uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	roleID, err := httpx.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - assign - role uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "Role assigned successfully") {
		return
	}

	err = c.u.Authz.Role.Assign(ctx.Request.Context(), userID, roleID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
