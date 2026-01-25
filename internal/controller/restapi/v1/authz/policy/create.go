package policy

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new policy
// @Description Create a new ABAC policy
// @Tags        authz-policies
// @Accept      json
// @Produce     json
// @Param       request body domain.Policy true "Policy creation body"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/policies [post]
func (c *Controller) Create(ctx *gin.Context) {
	var policy domain.Policy
	if err := ctx.ShouldBindJSON(&policy); err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - policy - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "Policy created successfully") {
		return
	}

	err := c.u.Authz.Policy.Create(ctx.Request.Context(), &policy)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
