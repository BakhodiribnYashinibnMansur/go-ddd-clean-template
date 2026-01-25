package errorcode

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	repo "gct/internal/repo/persistent/postgres/errorcode"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary     Update an error code
// @Description Update an existing error code definition
// @Tags        error_code
// @Accept      json
// @Produce     json
// @Param       code path string true "Error Code string"
// @Param       request body repo.UpdateErrorCodeInput true "Error code update input"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /error-codes/{code} [put]
func (c *Controller) Update(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "code is required", nil, false)
		return
	}

	var input repo.UpdateErrorCodeInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		c.logger.Error("fast http - v1 - errorcode - update - bind", "error", err)
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	ec, err := c.useCase.Update(ctx.Request.Context(), code, input)
	if err != nil {
		c.logger.Error("fast http - v1 - errorcode - update - service", "error", err)
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "error code updated", ec, true)
}
