package auth

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// SignIn godoc
// @Summary     Sign In
// @Description Authenticate user and return tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "Credentials"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-in [post]
func (c *Controller) SignIn(ctx *gin.Context) {
	var in domain.SignInIn
	if err := ctx.ShouldBindJSON(&in); err != nil {
		httpx.LogError(c.l, err, "http - v1 - auth - signin - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.SignInOut() }) {
		return
	}

	// Populate session info from request context
	c.populateSessionInfo(ctx, &in.Session)

	out, err := c.u.User.Client.SignIn(ctx.Request.Context(), &in)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusUnauthorized)
		return
	}

	// Set Cookies
	c.setAuthCookies(ctx, out)

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
