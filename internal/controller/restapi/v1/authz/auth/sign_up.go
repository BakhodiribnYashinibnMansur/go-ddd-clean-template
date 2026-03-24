package auth

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// SignUp godoc
// @Summary     Sign Up
// @Description Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.SignUpIn true "User info"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Validation error"
// @Failure     409 {object} response.ErrorResponse "Conflict"
// @Failure     500 {object} response.ErrorResponse "Internal error"
// @Router      /auth/sign-up [post]
func (c *Controller) SignUp(ctx *gin.Context) {
	var in domain.SignUpIn
	if err := ctx.ShouldBindJSON(&in); err != nil {
		httpx.LogError(c.l, err, "http - v1 - auth - signup - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.SignInOut() }) {
		return
	}

	// Populate session info from request context
	c.populateSessionInfo(ctx, in.Session)

	out, err := c.u.User.Client().SignUp(ctx.Request.Context(), &in)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	// Set Cookies
	c.setAuthCookies(ctx, out)

	response.ControllerResponse(ctx, http.StatusCreated, out, nil, true)
}
