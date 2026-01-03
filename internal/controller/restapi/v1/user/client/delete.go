package client

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"
)

// Delete godoc
// @Summary     Delete user
// @Description Delete user by ID (soft delete)
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{user_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := util.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		util.LogError(c.l, err, "http - v1 - client - delete - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	err = c.u.User.Client.Delete(ctx.Request.Context(), &domain.UserFilter{ID: &id})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
