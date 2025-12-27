package client

import (
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *Controller) Get(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	user, err := c.u.User.Client.GetByID(ctx.Request.Context(), id)
	if err != nil {
		c.l.Errorw("restapi - v1 - user - get", zap.Error(err))
		response.ControllerResponse(ctx, http.StatusInternalServerError, "service problems", nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, user, nil, true)
}

func (c *Controller) GetUsers(ctx *gin.Context) {
	// Not implemented in usecase yet, but for skeleton
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}
