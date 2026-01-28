package permission

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete permission
// @Description Delete permission by ID
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       perm_id path string true "Permission ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/permissions/{perm_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - permission - delete - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Permission deleted successfully") {
		return
	}

	err = c.u.Authz.Permission().Delete(ctx.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
