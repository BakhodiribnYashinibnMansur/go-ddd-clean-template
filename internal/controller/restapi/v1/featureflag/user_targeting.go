package featureflag

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/featureflag"

	"github.com/gin-gonic/gin"
)

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
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /featureflag/targeting [get]
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

	results := map[string]any{
		"enable-new-ui":    featureflag.IsEnabled(ctx, "enable-new-ui", false),
		"api-rate-limit":   featureflag.GetIntVariation(ctx, "api-rate-limit", 100),
		"homepage-variant": featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a"),
	}

	response.ControllerResponse(c, http.StatusOK, gin.H{
		"context_user": gin.H{
			"id":    userID,
			"email": email,
			"plan":  plan,
		},
		"evaluated_flags": results,
	}, nil, true)
}
