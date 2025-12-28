package session

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/gin-gonic/gin"
)

// UpdateActivity godoc
// @Summary     Update session activity
// @Description Update last activity timestamp
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id      path string true "Session UUID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions/{id}/activity [put]
func (c *Controller) UpdateActivity(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, "id")
	if err != nil {
		util.LogError(c.l, err, "http - v1 - session - updateActivity - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid session id", nil, false)
		return
	}

	filter := &domain.SessionFilter{ID: id}
	err = c.s.User.Session.UpdateActivity(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
