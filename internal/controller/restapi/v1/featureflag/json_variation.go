package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/featureflag"

	"github.com/gin-gonic/gin"
)

// ExampleJSONVariation demonstrates complex structured configuration.
// Allows pushing dictionaries or lists as feature configurations without hardcoding them.
// @Summary Example JSON Variation
// @Description Demonstrates JSON flags for shipping complex structured configurations
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/json [get]
func (ctrl *FeatureFlagController) ExampleJSONVariation(c *gin.Context) {
	ctx := c.Request.Context()

	defaultConfig := map[string]any{
		"maxItems":    10,
		"enableCache": true,
		"timeout":     30,
	}

	// retrieves a full JSON object mapped to the specified key.
	config := featureflag.GetJSONVariation(ctx, "feature-config", defaultConfig)

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"flag":   "feature-config",
		"config": config,
	}, nil, true)
}
