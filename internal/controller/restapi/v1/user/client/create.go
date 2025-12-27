package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *Controller) Create(ctx *gin.Context) {
	var body domain.User
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.l.Errorw("restapi - v1 - user - create", zap.Error(err))
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	err := c.u.User.Client.Create(ctx.Request.Context(), body)
	if err != nil {
		c.l.Errorw("restapi - v1 - user - create", zap.Error(err))
		response.ControllerResponse(ctx, http.StatusInternalServerError, "service problems", nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
