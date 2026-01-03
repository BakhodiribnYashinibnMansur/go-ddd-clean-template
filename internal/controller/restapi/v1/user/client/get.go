package client

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Get godoc
// @Summary     Get user by ID
// @Description Retrieve a user's details by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id  path string true "User ID"
// @Success     200 {object} response.SuccessResponse{data=domain.User}
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{user_id} [get]
func (c *Controller) User(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - client - get - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}
	out, err := c.u.User.Client.Get(ctx.Request.Context(), &domain.UserFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, out, nil, true)
}
