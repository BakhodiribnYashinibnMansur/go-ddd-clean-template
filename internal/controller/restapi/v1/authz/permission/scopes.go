package permission

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"github.com/gin-gonic/gin"
)

type ScopeRequest struct {
	Path   string `binding:"required" json:"path"`
	Method string `binding:"required" json:"method"`
}

// AssignScope godoc
// @Summary     Assign scope to permission
// @Description Assign a scope (path + method) to a permission
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       perm_id path string true "Permission ID"
// @Param       request body ScopeRequest true "Scope details"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/permissions/{perm_id}/scopes [post]
func (c *Controller) AssignScope(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - assign_scope - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	var req ScopeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - assign_scope - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeUpdate, "Scope assigned successfully") {
		return
	}

	err = c.u.Authz.Permission.AssignScope(ctx.Request.Context(), id, req.Path, req.Method)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}

// RemoveScope godoc
// @Summary     Remove scope from permission
// @Description Remove a scope (path + method) from a permission
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       perm_id path string true "Permission ID"
// @Param       request body ScopeRequest true "Scope details"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/permissions/{perm_id}/scopes [delete]
func (c *Controller) RemoveScope(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - remove_scope - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	var req ScopeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - remove_scope - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeDelete, "Scope removed successfully") {
		return
	}

	err = c.u.Authz.Permission.RemoveScope(ctx.Request.Context(), id, req.Path, req.Method)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
