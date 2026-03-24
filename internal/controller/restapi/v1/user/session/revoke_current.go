package session

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// RevokeCurrent godoc
// @Summary     Revoke current session
// @Description Revoke the session that is currently being used
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/current [delete]
func (c *Controller) RevokeCurrent(ctx *gin.Context) {
	sid, err := httpx.GetCtxSessionID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session not found", nil, false)
		return
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Current session revoked successfully") {
		return
	}

	err = c.s.User.Session().Delete(ctx.Request.Context(), &domain.SessionFilter{ID: &sid})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "Current session revoked successfully", nil, true)
}
