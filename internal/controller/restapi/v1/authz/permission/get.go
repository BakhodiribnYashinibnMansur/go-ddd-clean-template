package permission

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get permission by ID
// @Description Get permission details by unique ID
// @Tags        authz-permissions
// @Accept      json
// @Produce     json
// @Param       perm_id path string true "Permission ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/permissions/{perm_id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamPermID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - permission - get - uuid")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid permission id", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.PermissionWithID(id) }) {
		return
	}

	perm, err := c.u.Authz.Permission.Get(ctx.Request.Context(), &domain.PermissionFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, perm, nil, true)
}
