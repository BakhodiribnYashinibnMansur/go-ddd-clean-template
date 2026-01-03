package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
)

// Delete godoc
// @Summary     Delete scope
// @Tags        authz-scopes
// @Param       path query string true "Path"
// @Param       method query string true "Method"
// @Router      /authz/scopes [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	path, err := util.GetStringQuery(ctx, consts.QueryPath)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - delete - path")
		response.ControllerResponse(ctx, http.StatusBadRequest, "path required", nil, false)
		return
	}
	method, err := util.GetStringQuery(ctx, consts.QueryMethod)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - delete - method")
		response.ControllerResponse(ctx, http.StatusBadRequest, "method required", nil, false)
		return
	}

	err = c.u.Authz.Scope.Delete(ctx.Request.Context(), path, method)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
