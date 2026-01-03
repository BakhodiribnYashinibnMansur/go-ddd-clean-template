package role

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
)

// AddPermission godoc
// @Summary     Add permission to role
// @Description Add a permission to a role
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Param       perm_id path string true "Permission ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles/{role_id}/permissions/{perm_id} [post]
func (c *Controller) AddPermission(ctx *gin.Context) {
	roleID, err := util.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - add_permission - role uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	permID, err := util.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - add_permission - perm uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	err = c.u.Authz.Role.AddPermission(ctx.Request.Context(), roleID, permID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}

// RemovePermission godoc
// @Summary     Remove permission from role
// @Description Remove a permission from a role
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Param       perm_id path string true "Permission ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles/{role_id}/permissions/{perm_id} [delete]
func (c *Controller) RemovePermission(ctx *gin.Context) {
	roleID, err := util.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - remove_permission - role uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	permID, err := util.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - remove_permission - perm uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	err = c.u.Authz.Role.RemovePermission(ctx.Request.Context(), roleID, permID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
