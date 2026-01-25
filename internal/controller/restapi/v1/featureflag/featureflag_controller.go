// Package featureflag provides demonstration and interface for dynamic feature toggles.
// It allows for A/B testing, gradual rollouts, and runtime configuration changes without redeployment.
package featureflag

import (
	"net/http"

	"gct/pkg/featureflag"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// FeatureFlagController serves as a playground and reference implementation for the feature flag system.
type FeatureFlagController struct {
	logger logger.Log
}

// NewFeatureFlagController instantiates the controller with a structured logger.
func NewFeatureFlagController(logger logger.Log) *FeatureFlagController {
	return &FeatureFlagController{
		logger: logger,
	}
}

// ExampleBooleanFlag demonstrates a binary (True/False) toggle.
// This is typically used for enabling/disabling new code blocks or UI components.
// @Summary Example Boolean Flag
// @Description Demonstrates how to use a boolean feature flag to toggle UI elements
// @Tags feature-flags
// @Accept json
// @Produce json
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/boolean [get]
func (ctrl *FeatureFlagController) ExampleBooleanFlag(c *gin.Context) {
	ctx := c.Request.Context()

	// "enable-new-ui" is the key defined in the feature flag provider (e.g. GOFF or LaunchDarkly).
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-ui", false)

	c.JSON(http.StatusOK, gin.H{
		"flag":    "enable-new-ui",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "New UI is enabled"
			}
			return "Using legacy UI"
		}(),
	})
}

// ExampleStringVariation demonstrates multi-variant testing (A/B testing).
// Useful for comparing different themes, copy, or algorithmic approaches.
// @Summary Example String Variation
// @Description Demonstrates string-based flags for multi-variant A/B testing
// @Tags feature-flags
// @Accept json
// @Produce json
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/string [get]
func (ctrl *FeatureFlagController) ExampleStringVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// "homepage-variant" returns "variant-a", "variant-b", etc.
	variant := featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a")

	c.JSON(http.StatusOK, gin.H{
		"flag":    "homepage-variant",
		"variant": variant,
		"message": "User assigned to: " + variant,
	})
}

// ExampleIntVariation demonstrates remote numeric configuration.
// Often used for tuning rate limits, timeouts, or batch sizes at runtime.
// @Summary Example Int Variation
// @Description Demonstrates integer flags for dynamic numeric tuning (e.g. rate limits)
// @Tags feature-flags
// @Accept json
// @Produce json
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/int [get]
func (ctrl *FeatureFlagController) ExampleIntVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// "api-rate-limit" can be increased or decreased globally based on server load.
	rateLimit := featureflag.GetIntVariation(ctx, "api-rate-limit", 100)

	c.JSON(http.StatusOK, gin.H{
		"flag":      "api-rate-limit",
		"rateLimit": rateLimit,
		"message":   "Current effective rate limit",
		"unit":      "requests per minute",
	})
}

// ExampleJSONVariation demonstrates complex structured configuration.
// Allows pushing dictionaries or lists as feature configurations without hardcoding them.
// @Summary Example JSON Variation
// @Description Demonstrates JSON flags for shipping complex structured configurations
// @Tags feature-flags
// @Accept json
// @Produce json
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/json [get]
func (ctrl *FeatureFlagController) ExampleJSONVariation(c *gin.Context) {
	ctx := c.Request.Context()

	defaultConfig := map[string]interface{}{
		"maxItems":    10,
		"enableCache": true,
		"timeout":     30,
	}

	// retrieves a full JSON object mapped to the specified key.
	config := featureflag.GetJSONVariation(ctx, "feature-config", defaultConfig)

	c.JSON(http.StatusOK, gin.H{
		"flag":   "feature-config",
		"config": config,
	})
}

// ExampleUserTargeting demonstrates segmenting users based on attributes.
// Flags can be targeted to specific internal users, beta testers, or premium subscribers.
// @Summary Example User Targeting
// @Description Demonstrates segment-based flag targeting (e.g. per plan or email)
// @Tags feature-flags
// @Accept json
// @Produce json
// @Param user_id query string false "Unique identity for targeting"
// @Param email query string false "Email address for precise targeting"
// @Param plan query string false "Segment identifier (free/premium)"
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/targeting [get]
func (ctrl *FeatureFlagController) ExampleUserTargeting(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.DefaultQuery("user_id", "anonymous")
	email := c.DefaultQuery("email", "")
	plan := c.DefaultQuery("plan", "free")

	// Map current request context to an identifiable feature-flag User.
	var user featureflag.User
	if userID == "anonymous" {
		user = featureflag.NewAnonymousUser()
	} else {
		user = featureflag.NewUser(userID)
	}

	if email != "" {
		user = user.WithEmail(email)
	}
	user = user.WithCustom("plan", plan)

	// Embed user metadata into context for the evaluation engine.
	ctx = featureflag.WithUser(ctx, user)

	results := map[string]interface{}{
		"enable-new-ui":    featureflag.IsEnabled(ctx, "enable-new-ui", false),
		"api-rate-limit":   featureflag.GetIntVariation(ctx, "api-rate-limit", 100),
		"homepage-variant": featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a"),
	}

	c.JSON(http.StatusOK, gin.H{
		"context_user": gin.H{
			"id":    userID,
			"email": email,
			"plan":  plan,
		},
		"evaluated_flags": results,
	})
}

// ExamplePercentageRollout demonstrates canary/percentage-based deployments.
// @Summary Example Percentage Rollout
// @Description Demonstrates gradual percentage-based rollouts to minimize risk
// @Tags feature-flags
// @Accept json
// @Produce json
// @Failure 500 {object} response.SuccessResponse
// @Security BearerAuth
// @Router /api/v1/featureflag/rollout [get]
func (ctrl *FeatureFlagController) ExamplePercentageRollout(c *gin.Context) {
	ctx := c.Request.Context()

	// Evaluates whether the current (likely randomized) user falls into the rollout bucket.
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-feature", false)

	c.JSON(http.StatusOK, gin.H{
		"flag":    "enable-new-feature",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "You have been selected for this rollout"
			}
			return "This feature is not yet available for your segment"
		}(),
	})
}
