package role

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Update godoc
// @Summary     Update role
// @Description Update role name
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Param       request body domain.Role true "Role update body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /authz/roles/{role_id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - update - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	var role domain.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - role - update - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	role.ID = id

	err = c.u.Authz.Role.Update(ctx.Request.Context(), &role)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
