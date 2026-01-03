package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Create godoc
// @Summary     Create a new scope
// @Tags        authz-scopes
// @Accept      json
// @Produce     json
// @Param       request body domain.Scope true "Scope creation body"
// @Router      /authz/scopes [post]
func (c *Controller) Create(ctx *gin.Context) {
	var scope domain.Scope
	if err := ctx.ShouldBindJSON(&scope); err != nil {
		util.LogError(c.l, err, "http - v1 - authz - scope - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	err := c.u.Authz.Scope.Create(ctx.Request.Context(), &scope)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
