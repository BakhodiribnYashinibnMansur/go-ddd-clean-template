package client

import (
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	var body domain.User
	if err := ctx.ShouldBindJSON(&body); err != nil {
		c.l.Errorw("restapi - v1 - user - update", zap.Error(err))
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	body.ID = id

	err = c.u.User.Client.Update(ctx.Request.Context(), body)
	if err != nil {
		c.l.Errorw("restapi - v1 - user - update", zap.Error(err))
		response.ControllerResponse(ctx, http.StatusInternalServerError, "service problems", nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
