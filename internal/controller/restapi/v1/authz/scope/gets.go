package scope

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     List scopes
// @Tags        authz-scopes
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       path query string false "Path"
// @Param       method query string false "Method"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/scopes [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	path := httpx.GetNullStringQuery(ctx, consts.QueryPath)
	method := httpx.GetNullStringQuery(ctx, consts.QueryMethod)

	filter := domain.ScopesFilter{
		Pagination: &pagination,
	}
	if path != "" {
		filter.Path = &path
	}
	if method != "" {
		filter.Method = &method
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Scopes(count) }) {
		return
	}

	scopes, count, err := c.u.Authz.Scope().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, scopes, count, true)
}
