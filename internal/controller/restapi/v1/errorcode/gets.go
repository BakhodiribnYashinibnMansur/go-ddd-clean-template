package errorcode

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

// List godoc
// @Summary     Get all error codes
// @Description Retrieve all error code definitions
// @Tags        error_code
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /error-codes [get]
func (c *Controller) Gets(ctx *gin.Context) {
	ecs, err := c.useCase.List(ctx.Request.Context())
	if err != nil {
		c.logger.Error("fast http - v1 - errorcode - gets - service", "error", err)
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, ecs, nil, true)
}
