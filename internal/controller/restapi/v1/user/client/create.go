package client

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new user
// @Description Register a new user with username, phone and password
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "User creation query"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /users [post]
// @Auth
func (c *Controller) Create(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "User created successfully") {
		return
	}

	err := c.u.User.Client.Create(ctx.Request.Context(), &user)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
