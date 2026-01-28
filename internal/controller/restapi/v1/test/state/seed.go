package state

import (
	"gct/internal/controller/restapi/response"
	apperrors "gct/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Seed godoc
// @Summary Seed database with custom counts
// @Description Seeds database with specified counts. Only available in dev/test.
// @Tags Test
// @Accept json
// @Produce json
// @Param request body map[string]int true "Seed counts (users, roles, permissions, policies)"
// @Success 200 {object} response.SuccessResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /api/v1/test/seed [post]
func (c *Controller) Seed(ctx *gin.Context) {
	if c.cfg.App.IsProd() {
		response.ControllerResponse(ctx, http.StatusForbidden, apperrors.ErrForbidden, nil, false)
		return
	}

	var customCounts map[string]int
	if err := ctx.ShouldBindJSON(&customCounts); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	if err := c.seeder.Seed(ctx.Request.Context(), customCounts); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, "Database seeded successfully", nil, true)
}
