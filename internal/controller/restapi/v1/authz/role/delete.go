package role

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
)

// Delete godoc
// @Summary     Delete role
// @Description Delete role by ID
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles/{role_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - delete - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	err = c.u.Authz.Role.Delete(ctx.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
