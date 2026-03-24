package client

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new user
// @Description Register a new user with username, phone and password. Requires admin privileges.
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "User creation data (username, phone, password required)"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid request body, validation failed, or duplicate user"
// @Failure     401 {object} response.ErrorResponse "Unauthorized - missing or invalid token"
// @Failure     403 {object} response.ErrorResponse "Forbidden - insufficient permissions"
// @Failure     500 {object} response.ErrorResponse "Internal server error"
// @Security    BearerAuth
// @Router      /users [post]
func (c *Controller) Create(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - create - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeCreate, "User created successfully") {
		return
	}

	err := c.u.User.Client().Create(ctx.Request.Context(), &user)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
