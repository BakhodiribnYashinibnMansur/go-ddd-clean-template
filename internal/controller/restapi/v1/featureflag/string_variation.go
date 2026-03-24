package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/featureflag"

	"github.com/gin-gonic/gin"
)

// ExampleStringVariation demonstrates multi-variant testing (A/B testing).
// Useful for comparing different themes, copy, or algorithmic approaches.
// @Summary Example String Variation
// @Description Demonstrates string-based flags for multi-variant A/B testing
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/string [get]
func (ctrl *FeatureFlagController) ExampleStringVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// "homepage-variant" returns "variant-a", "variant-b", etc.
	variant := featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a")

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"flag":       "homepage-variant",
		"variation":  variant,
		"message":    "Dynamic content variation",
		"assignment": "based on targeting rules",
	}, nil, true)
}
