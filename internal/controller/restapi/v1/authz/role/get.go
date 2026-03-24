package role

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get role by ID
// @Description Get role details by unique ID
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       role_id path string true "Role ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/roles/{role_id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamRoleID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - get - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid role id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.RoleWithID(id) }) {
		return
	}

	role, err := c.u.Authz.Role().Get(ctx.Request.Context(), &domain.RoleFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, role, nil, true)
}
