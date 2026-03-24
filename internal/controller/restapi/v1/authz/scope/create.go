package scope

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new scope
// @Tags        authz-scopes
// @Accept      json
// @Produce     json
// @Param       request body domain.Scope true "Scope creation body"
// @Success     201 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/scopes [post]
func (c *Controller) Create(ctx *gin.Context) {
	var scope domain.Scope
	if err := ctx.ShouldBindJSON(&scope); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - scope - create - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "Scope created successfully") {
		return
	}

	err := c.u.Authz.Scope().Create(ctx.Request.Context(), &scope)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
