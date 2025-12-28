package session

import (
	"net/http"

	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RevokeAll godoc
// @Summary     Revoke all other sessions
// @Description Revoke all user sessions except the current one (Currently revokes current one as placeholder)
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions/revoke-all [post]
func (c *Controller) RevokeAll(ctx *gin.Context) {
	userID, exists := ctx.Get(consts.CtxUserID)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	currentSessionID, exists := ctx.Get(consts.CtxSessionID)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "session not found", nil, false)
		return
	}

	// Logging purpose only
	uid, _ := userID.(string)

	sid, ok := currentSessionID.(string)
	if !ok {
		util.LogError(c.l, nil, "http - v1 - session - revoke_all - context_sid: invalid type")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "invalid session ID", nil, false)
		return
	}

	sessionUUID, err := uuid.Parse(sid)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - session - revoke_all - uuid")
		response.ControllerResponse(ctx, http.StatusInternalServerError, "invalid session ID format", nil, false)
		return
	}

	// Just revoke current session for now
	filter := &domain.SessionFilter{ID: sessionUUID}
	err = c.s.User.Session.Revoke(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	c.l.Infow("Session revoked for user", "user_id", uid)
	response.ControllerResponse(ctx, http.StatusOK, "Session revoked successfully", nil, true)
}
