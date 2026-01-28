package errorcode

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

// Get godoc
// @Summary     Get an error code
// @Description Get an error code details by code string
// @Tags        error_code
// @Accept      json
// @Produce     json
// @Param       code path string true "Error Code string"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /error-codes/{code} [get]
func (c *Controller) Get(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "code is required", nil, false)
		return
	}

	ec, err := c.useCase.GetByCode(ctx.Request.Context(), code)
	if err != nil {
		c.logger.Error("fast http - v1 - errorcode - get - service", "error", err)
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "error code found", ec, true)
}
