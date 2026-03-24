package role

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     List roles
// @Description Get list of roles with pagination and optional filtering
// @Tags        authz-roles
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       name query string false "Filter by name"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/roles [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - role - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	roleName := httpx.GetNullStringQuery(ctx, consts.QueryName)
	filter := domain.RolesFilter{
		Pagination: &pagination,
	}
	if roleName != "" {
		filter.Name = &roleName
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Roles(count) }) {
		return
	}

	roles, count, err := c.u.Authz.Role().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, roles, count, true)
}
