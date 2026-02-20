package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/featureflag"

	"github.com/gin-gonic/gin"
)

// ExamplePercentageRollout demonstrates canary/percentage-based deployments.
// @Summary Example Percentage Rollout
// @Description Demonstrates gradual percentage-based rollouts to minimize risk
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/rollout [get]
func (ctrl *FeatureFlagController) ExamplePercentageRollout(c *gin.Context) {
	ctx := c.Request.Context()

	// Evaluates whether the current (likely randomized) user falls into the rollout bucket.
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-feature", false)

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"flag":    "enable-new-feature",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "You have been selected for this rollout"
			}
			return "This feature is not yet available for your segment"
		}(),
	}, nil, true)
}
