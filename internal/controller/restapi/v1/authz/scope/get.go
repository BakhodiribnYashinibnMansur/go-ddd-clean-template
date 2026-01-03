package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Get godoc
// @Summary     Get scope
// @Tags        authz-scopes
// @Param       path query string true "Path"
// @Param       method query string true "Method"
// @Router      /authz/scope [get]
func (c *Controller) Get(ctx *gin.Context) {
	path, err := util.GetStringQuery(ctx, consts.QueryPath)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - get - path")
		response.ControllerResponse(ctx, http.StatusBadRequest, "path required", nil, false)
		return
	}
	method, err := util.GetStringQuery(ctx, consts.QueryMethod)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - get - method")
		response.ControllerResponse(ctx, http.StatusBadRequest, "method required", nil, false)
		return
	}

	scope, err := c.u.Authz.Scope.Get(ctx.Request.Context(), &domain.ScopeFilter{Path: &path, Method: &method})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, scope, nil, true)
}
