package role

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Get godoc
// @Summary     Get role by ID
// @Description Get role details by unique ID
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles/{role_id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - get - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	role, err := c.u.Authz.Role.Get(ctx.Request.Context(), &domain.RoleFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, role, nil, true)
}
