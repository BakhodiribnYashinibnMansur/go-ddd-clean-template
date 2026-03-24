package role

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
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
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/roles/{role_id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - update - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

	var role domain.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - update - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	role.ID = id
	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "Role updated successfully") {
		return
	}

	err = c.u.Authz.Role().Update(ctx.Request.Context(), &role)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
