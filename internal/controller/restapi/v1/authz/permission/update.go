package permission

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary     Update permission
// @Description Update permission details
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       perm_id path string true "Permission ID"
// @Param       request body domain.Permission true "Permission update body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/permissions/{perm_id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - update - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

	var perm domain.Permission
	if err := ctx.ShouldBindJSON(&perm); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - permission - update - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	perm.ID = id
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeUpdate, "Permission updated successfully") {
		return
	}

	err = c.u.Authz.Permission.Update(ctx.Request.Context(), &perm)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
