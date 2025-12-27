package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (c *Controller) UpdateActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(ctx, http.StatusBadRequest, "invalid session id")
		return
	}

	type updateRequest struct {
		FCMToken *string `json:"fcm_token"`
	}
	var body updateRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.l.Errorw("restapi - v1 - session - updateActivity", zap.Error(err))
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	err = c.s.User.Session.UpdateActivity(ctx.Request.Context(), id)
	if err != nil {
		c.l.Errorw("restapi - v1 - session - updateActivity", zap.Error(err))
		errorResponse(ctx, http.StatusInternalServerError, "service problems")
		return
	}
	ctx.Status(http.StatusOK)
}
