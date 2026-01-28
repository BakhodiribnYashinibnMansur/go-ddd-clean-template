package auth

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// RefreshToken godoc
// @Summary     Refresh tokens
// @Description Rotates access and refresh tokens using a valid refresh token.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.RefreshIn true "Refresh request body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /auth/refresh [post]
func (c *Controller) RefreshToken(ctx *gin.Context) {
	var req domain.RefreshIn
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpx.LogError(c.l, err, "http - v1 - auth - refresh - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	// Rotate session
	out, err := c.u.User.Client().RotateSession(ctx.Request.Context(), &req)
	if err != nil {
		// Security: Always return 401 for refresh failures
		response.ControllerResponse(ctx, http.StatusUnauthorized, "invalid session", nil, false)
		return
	}

	// Set new Cookies
	c.setAuthCookies(ctx, out)

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
