package client

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Update godoc
// @Summary     Update user
// @Description Update user details by ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path int                       true "User ID"
// @Param       request body domain.User true "User update query"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{user_id} [patch]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := util.GetInt64Param(ctx, consts.ParamUserID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - client - update - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		util.LogError(c.l, err, "http - v1 - client - update - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}
	user.ID = id

	err = c.u.User.Client.Update(ctx.Request.Context(), &user)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
