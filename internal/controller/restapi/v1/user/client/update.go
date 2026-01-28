package client

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary     Update user
// @Description Update user details by ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "User ID"
// @Param       request body domain.User true "User update query"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /users/{user_id} [patch]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := httpx.GetUUIDParam(ctx, consts.ParamUserID)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - update - id")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		httpx.LogError(c.l, err, "http - v1 - client - update - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	user.ID = id

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeUpdate, "User updated successfully") {
		return
	}

	err = c.u.User.Client().Update(ctx.Request.Context(), &user)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
