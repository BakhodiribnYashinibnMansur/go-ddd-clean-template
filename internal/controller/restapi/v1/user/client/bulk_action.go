package client

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

// BulkAction godoc
// @Summary     Bulk action on users
// @Description Deactivate or delete multiple users by IDs
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       body body domain.BulkActionRequest true "Bulk action request"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse "Invalid request body"
// @Failure     401 {object} response.ErrorResponse "Unauthorized"
// @Failure     403 {object} response.ErrorResponse "Forbidden"
// @Failure     500 {object} response.ErrorResponse "Internal server error"
// @Security    BearerAuth
// @Router      /users/bulk-action [post]
func (c *Controller) BulkAction(ctx *gin.Context) {
	var req domain.BulkActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	if err := c.u.User.Client().BulkAction(ctx.Request.Context(), req); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
