package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (c *Controller) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errorResponse(ctx, http.StatusBadRequest, "invalid session id")
		return
	}

	session, err := c.s.User.Session.GetByID(ctx.Request.Context(), id)
	if err != nil {
		c.l.Errorw("restapi - v1 - session - get", zap.Error(err))
		errorResponse(ctx, http.StatusInternalServerError, "service problems")
		return
	}

	ctx.JSON(http.StatusOK, session)
}
