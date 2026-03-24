package scope

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete scope
// @Tags        authz-scopes
// @Param       path query string true "Path"
// @Param       method query string true "Method"
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     401,403,400,500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/scopes [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	path, err := httpx.GetStringQuery(ctx, consts.QueryPath)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - delete - path")
		response.ControllerResponse(ctx, http.StatusBadRequest, "path required", nil, false)
		return
	}
	method, err := httpx.GetStringQuery(ctx, consts.QueryMethod)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - delete - method")
		response.ControllerResponse(ctx, http.StatusBadRequest, "method required", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Scope deleted successfully") {
		return
	}

	err = c.u.Authz.Scope().Delete(ctx.Request.Context(), path, method)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
