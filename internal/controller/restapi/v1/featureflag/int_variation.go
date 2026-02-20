package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/featureflag"

	"github.com/gin-gonic/gin"
)

// ExampleIntVariation demonstrates remote numeric configuration.
// Often used for tuning rate limits, timeouts, or batch sizes at runtime.
// @Summary Example Int Variation
// @Description Demonstrates integer flags for dynamic numeric tuning (e.g. rate limits)
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/int [get]
func (ctrl *FeatureFlagController) ExampleIntVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// "api-rate-limit" can be increased or decreased globally based on server load.
	rateLimit := featureflag.GetIntVariation(ctx, "api-rate-limit", 100)

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"flag":      "api-rate-limit",
		"rateLimit": rateLimit,
		"message":   "Current effective rate limit",
		"unit":      "requests per minute",
	}, nil, true)
}
