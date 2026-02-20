package role

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// RemovePermission godoc
// @Summary     Remove permission from role
// @Description Remove a permission from a role
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Param       perm_id path string true "Permission ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/roles/{role_id}/permissions/{perm_id} [delete]
func (c *Controller) RemovePermission(ctx *gin.Context) {
	roleID, err := httpx.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - remove_permission - role uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	permID, err := httpx.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - remove_permission - perm uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Permission removed from role successfully") {
		return
	}

	err = c.u.Authz.Role().RemovePermission(ctx.Request.Context(), roleID, permID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
