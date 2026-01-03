package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Gets godoc
// @Summary     List scopes
// @Tags        authz-scopes
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       path query string false "Path"
// @Param       method query string false "Method"
// @Router      /authz/scopes [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := util.GetPagination(ctx)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	path := util.GetNullStringQuery(ctx, consts.QueryPath)
	method := util.GetNullStringQuery(ctx, consts.QueryMethod)

	filter := domain.ScopesFilter{
		Pagination: &pagination,
	}
	if path != "" {
		filter.Path = &path
	}
	if method != "" {
		filter.Method = &method
	}

	scopes, count, err := c.u.Authz.Scope.Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, scopes, count, true)
}
