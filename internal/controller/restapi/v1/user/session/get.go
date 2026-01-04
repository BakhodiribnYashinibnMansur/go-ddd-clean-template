package session

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get session by ID
// @Description Retrieve a session details
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id  path string true "Session UUID"
// @Success     200 {object} response.SuccessResponse{data=domain.Session}
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions/{id} [get]
func (c *Controller) Session(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - session - get - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid session id", nil, false)
		return
	}

	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return mock.Session() }) {
		return
	}

	filter := &domain.SessionFilter{ID: &id}
	session, err := c.s.User.Session.Get(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, session, nil, true)
}
