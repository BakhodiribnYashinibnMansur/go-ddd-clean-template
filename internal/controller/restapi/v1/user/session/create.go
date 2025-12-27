package session

import (
	"net/http"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *Controller) Create(ctx *gin.Context) {
	var body domain.Session

	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.l.Errorw("restapi - v1 - session - create", zap.Error(err))
		errorResponse(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	// Default duration could also be part of config or request
	duration := 24 * time.Hour

	createdSession, err := c.s.User.Session.Create(ctx.Request.Context(), body, duration)
	if err != nil {
		c.l.Errorw("restapi - v1 - session - create", zap.Error(err))
		errorResponse(ctx, http.StatusInternalServerError, "service problems")
		return
	}

	ctx.JSON(http.StatusCreated, createdSession)
}

func errorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, response.Error{Error: msg})
}
