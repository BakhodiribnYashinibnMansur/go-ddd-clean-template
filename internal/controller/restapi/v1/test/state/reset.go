package state

import (
	"gct/internal/controller/restapi/response"
	apperrors "gct/internal/shared/infrastructure/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Reset godoc
// @Summary Reset database
// @Description Truncates all tables and re-seeds with base data. Only available in dev/test.
// @Tags Test
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /api/v1/test/reset [post]
func (c *Controller) Reset(ctx *gin.Context) {
	if c.cfg.App.IsProd() {
		response.ControllerResponse(ctx, http.StatusForbidden, apperrors.ErrForbidden, nil, false)
		return
	}

	customCounts := map[string]int{
		"clear_data": 1,
		"seed":       1337, // Fixed seed for tests
	}

	if err := c.seeder.Seed(ctx.Request.Context(), customCounts); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "Database reset and re-seeded successfully", nil, true)
}
