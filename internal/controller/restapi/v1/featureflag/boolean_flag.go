package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/featureflag"

	"github.com/gin-gonic/gin"
)

// ExampleBooleanFlag demonstrates a binary (True/False) toggle.
// This is typically used for enabling/disabling new code blocks or UI components.
// @Summary Example Boolean Flag
// @Description Demonstrates how to use a boolean feature flag to toggle UI elements
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/boolean [get]
func (ctrl *FeatureFlagController) ExampleBooleanFlag(c *gin.Context) {
	ctx := c.Request.Context()

	// "enable-new-ui" is the key defined in the feature flag provider (e.g. GOFF or LaunchDarkly).
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-ui", false)

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"flag":    "enable-new-ui",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "New UI is enabled"
			}
			return "Using legacy UI"
		}(),
	}, nil, true)
}
