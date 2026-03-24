package relation

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new relation
// @Description Create a new organizational relation
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       request body domain.Relation true "Relation creation body"
// @Success     201 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations [post]
func (c *Controller) Create(ctx *gin.Context) {
	var relation domain.Relation
	if err := ctx.ShouldBindJSON(&relation); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - create - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "Relation created successfully") {
		return
	}

	err := c.u.Authz.Relation().Create(ctx.Request.Context(), &relation)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
