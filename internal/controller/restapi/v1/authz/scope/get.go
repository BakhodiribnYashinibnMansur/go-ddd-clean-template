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

// Get godoc
// @Summary     Get scope
// @Tags        authz-scopes
// @Param       path query string true "Path"
// @Param       method query string true "Method"
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/scope [get]
func (c *Controller) Get(ctx *gin.Context) {
	path, err := httpx.GetStringQuery(ctx, consts.QueryPath)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - get - path")
		response.ControllerResponse(ctx, http.StatusBadRequest, "path required", nil, false)
		return
	}
	method, err := httpx.GetStringQuery(ctx, consts.QueryMethod)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - get - method")
		response.ControllerResponse(ctx, http.StatusBadRequest, "method required", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.Scope() }) {
		return
	}

	scope, err := c.u.Authz.Scope().Get(ctx.Request.Context(), &domain.ScopeFilter{Path: &path, Method: &method})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, scope, nil, true)
}
