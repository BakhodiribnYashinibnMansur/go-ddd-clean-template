package session

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

// RevokeAll godoc
// @Summary     Revoke all other sessions
// @Description Revoke all user sessions except the current one
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       request body domain.RevokeSessionsIn true "Revoke sessions request body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /sessions/revoke-all [post]
func (c *Controller) RevokeAll(ctx *gin.Context) {
	userID, err := httpx.GetUserID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	sid, err := httpx.GetCtxSessionID(ctx)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session not found", nil, false)
		return
	}

	var req domain.RevokeSessionsIn
	// Body is optional — UserID comes from JWT context
	_ = ctx.ShouldBindJSON(&req)
	req.UserID = userID

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeDelete, "Sessions revoked successfully") {
		return
	}

	// Revoke all sessions for this user
	filter := &domain.SessionFilter{UserID: &userID}
	err = c.s.User.Session().Revoke(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	c.l.Infow("All sessions revoked for user", "user_id", userID, "current_session_id", sid)
	response.ControllerResponse(ctx, http.StatusOK, "All sessions revoked successfully", nil, true)
}
