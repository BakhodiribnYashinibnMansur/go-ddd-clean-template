package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get user by ID
// @Description Retrieve a user's details by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id  path int true "User ID"
// @Success     200 {object} response.SuccessResponse{data=domain.User}
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{user_id} [get]
func (c *Controller) User(ctx *gin.Context) {
	id, err := util.GetInt64Param(ctx, "user_id")
	if err != nil {
		util.LogError(c.l, err, "http - v1 - client - get - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}
	out, err := c.u.User.Client.User(ctx.Request.Context(), uc_client.UserInput{ID: id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, out.User, nil, true)
}
