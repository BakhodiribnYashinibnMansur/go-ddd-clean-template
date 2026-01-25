package errorcode

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	repo "gct/internal/repo/persistent/postgres/errorcode"

	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new error code
// @Description Create a new error code definition
// @Tags        error_code
// @Accept      json
// @Produce     json
// @Param       request body repo.CreateErrorCodeInput true "Error code creation input"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /error-codes [post]
func (c *Controller) Create(ctx *gin.Context) {
	var input repo.CreateErrorCodeInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		c.logger.Error("fast http - v1 - errorcode - create - bind", "error", err)
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	ec, err := c.useCase.Create(ctx.Request.Context(), input)
	if err != nil {
		c.logger.Error("fast http - v1 - errorcode - create - service", "error", err)
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, "error code created", ec, true)
}
