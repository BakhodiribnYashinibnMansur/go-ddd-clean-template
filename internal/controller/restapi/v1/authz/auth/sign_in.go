package auth

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// SignIn godoc
// @Summary     Sign In
// @Description Authenticate user with credentials and return access/refresh tokens
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.SignInIn true "User credentials (phone/email/username and password)"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid request"
// @Failure     401 {object} response.ErrorResponse "Unauthorized"
// @Failure     404 {object} response.ErrorResponse "User not found"
// @Failure     500 {object} response.ErrorResponse "Internal error"
// @Router      /auth/sign-in [post]
func (c *Controller) SignIn(ctx *gin.Context) {
	var in domain.SignInIn
	if err := ctx.ShouldBindJSON(&in); err != nil {
		httpx.LogError(c.l, err, "http - v1 - auth - signin - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.SignInOut() }) {
		return
	}

	// Populate session info from request context
	c.populateSessionInfo(ctx, in.Session)

	out, err := c.u.User.Client().SignIn(ctx.Request.Context(), &in)
	if err != nil {
		// Security: Always return 401 for sign-in failures to prevent enumeration
		// regardless of whether the user exists (404) or password was wrong (401).
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid credentials", nil, false)
		return
	}

	// Set Cookies
	c.setAuthCookies(ctx, out)

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
