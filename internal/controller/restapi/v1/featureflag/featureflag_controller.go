package featureflag

import (
	"net/http"

	"gct/pkg/featureflag"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FeatureFlagController demonstrates feature flag usage.
type FeatureFlagController struct {
	logger *zap.Logger
}

// NewFeatureFlagController creates a new feature flag controller.
func NewFeatureFlagController(logger *zap.Logger) *FeatureFlagController {
	return &FeatureFlagController{
		logger: logger,
	}
}

// ExampleBooleanFlag demonstrates a simple boolean feature flag.
// @Summary Example Boolean Flag
// @Description Demonstrates how to use a boolean feature flag
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/boolean [get]
func (ctrl *FeatureFlagController) ExampleBooleanFlag(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if new feature is enabled
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-ui", false)

	c.JSON(http.StatusOK, gin.H{
		"flag":    "enable-new-ui",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "New UI is enabled"
			}
			return "Using old UI"
		}(),
	})
}

// ExampleStringVariation demonstrates a string variation (A/B testing).
// @Summary Example String Variation
// @Description Demonstrates how to use string variation for A/B testing
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/string [get]
func (ctrl *FeatureFlagController) ExampleStringVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// Get homepage variant
	variant := featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a")

	c.JSON(http.StatusOK, gin.H{
		"flag":    "homepage-variant",
		"variant": variant,
		"message": "User will see " + variant,
	})
}

// ExampleIntVariation demonstrates an integer variation (rate limiting).
// @Summary Example Int Variation
// @Description Demonstrates how to use integer variation for rate limiting
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/int [get]
func (ctrl *FeatureFlagController) ExampleIntVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// Get rate limit
	rateLimit := featureflag.GetIntVariation(ctx, "api-rate-limit", 100)

	c.JSON(http.StatusOK, gin.H{
		"flag":      "api-rate-limit",
		"rateLimit": rateLimit,
		"message":   "User rate limit is set",
		"unit":      "requests per minute",
	})
}

// ExampleJSONVariation demonstrates a JSON variation (complex configuration).
// @Summary Example JSON Variation
// @Description Demonstrates how to use JSON variation for complex configuration
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/json [get]
func (ctrl *FeatureFlagController) ExampleJSONVariation(c *gin.Context) {
	ctx := c.Request.Context()

	// Default configuration
	defaultConfig := map[string]interface{}{
		"maxItems":    10,
		"enableCache": true,
		"timeout":     30,
	}

	// Get feature configuration
	config := featureflag.GetJSONVariation(ctx, "feature-config", defaultConfig)

	c.JSON(http.StatusOK, gin.H{
		"flag":   "feature-config",
		"config": config,
	})
}

// ExampleUserTargeting demonstrates user-specific targeting.
// @Summary Example User Targeting
// @Description Demonstrates how to use feature flags with user targeting
// @Tags feature-flags
// @Accept json
// @Produce json
// @Param user_id query string false "User ID"
// @Param email query string false "User Email"
// @Param plan query string false "User Plan (free/premium)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/targeting [get]
func (ctrl *FeatureFlagController) ExampleUserTargeting(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user information from query params
	userID := c.DefaultQuery("user_id", "anonymous")
	email := c.DefaultQuery("email", "")
	plan := c.DefaultQuery("plan", "free")

	// Create feature flag user
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

	// Add user to context
	ctx = featureflag.WithUser(ctx, user)

	// Check multiple flags
	results := map[string]interface{}{
		"enable-new-ui":    featureflag.IsEnabled(ctx, "enable-new-ui", false),
		"api-rate-limit":   featureflag.GetIntVariation(ctx, "api-rate-limit", 100),
		"homepage-variant": featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a"),
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"email": email,
			"plan":  plan,
		},
		"flags": results,
	})
}

// ExamplePercentageRollout demonstrates percentage-based rollout.
// @Summary Example Percentage Rollout
// @Description Demonstrates how feature flags can be rolled out to a percentage of users
// @Tags feature-flags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/featureflag/rollout [get]
func (ctrl *FeatureFlagController) ExamplePercentageRollout(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if new feature is enabled (with percentage rollout)
	isEnabled := featureflag.IsEnabled(ctx, "enable-new-feature", false)

	c.JSON(http.StatusOK, gin.H{
		"flag":    "enable-new-feature",
		"enabled": isEnabled,
		"message": func() string {
			if isEnabled {
				return "You are part of the rollout group"
			}
			return "Feature not yet available for you"
		}(),
	})
}
