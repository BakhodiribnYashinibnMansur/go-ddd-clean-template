package featureflag

import (
	"context"

	"github.com/gin-gonic/gin"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// FeatureFlagClientKey is the context key for the feature flag client.
	FeatureFlagClientKey contextKey = "feature_flag_client"
	// FeatureFlagUserKey is the context key for the feature flag user.
	FeatureFlagUserKey contextKey = "feature_flag_user"
)

// Middleware injects the feature flag client into the Gin context.
func Middleware(client *Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store the client in the context
		ctx := context.WithValue(c.Request.Context(), FeatureFlagClientKey, client)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// WithUser adds a user to the context for feature flag evaluation.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, FeatureFlagUserKey, user)
}

// GetClient retrieves the feature flag client from the context.
func GetClient(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(FeatureFlagClientKey).(*Client)
	return client, ok
}

// GetUser retrieves the feature flag user from the context.
func GetUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(FeatureFlagUserKey).(User)
	return user, ok
}

// IsEnabled is a helper function to check if a feature flag is enabled from context.
func IsEnabled(ctx context.Context, flagKey string, defaultValue bool) bool {
	client, ok := GetClient(ctx)
	if !ok {
		return defaultValue
	}

	user, ok := GetUser(ctx)
	if !ok {
		user = NewAnonymousUser()
	}

	return client.IsEnabled(ctx, flagKey, user, defaultValue)
}

// GetStringVariation is a helper function to get a string variation from context.
func GetStringVariation(ctx context.Context, flagKey string, defaultValue string) string {
	client, ok := GetClient(ctx)
	if !ok {
		return defaultValue
	}

	user, ok := GetUser(ctx)
	if !ok {
		user = NewAnonymousUser()
	}

	return client.GetStringVariation(ctx, flagKey, user, defaultValue)
}

// GetIntVariation is a helper function to get an int variation from context.
func GetIntVariation(ctx context.Context, flagKey string, defaultValue int) int {
	client, ok := GetClient(ctx)
	if !ok {
		return defaultValue
	}

	user, ok := GetUser(ctx)
	if !ok {
		user = NewAnonymousUser()
	}

	return client.GetIntVariation(ctx, flagKey, user, defaultValue)
}

// GetFloatVariation is a helper function to get a float variation from context.
func GetFloatVariation(ctx context.Context, flagKey string, defaultValue float64) float64 {
	client, ok := GetClient(ctx)
	if !ok {
		return defaultValue
	}

	user, ok := GetUser(ctx)
	if !ok {
		user = NewAnonymousUser()
	}

	return client.GetFloatVariation(ctx, flagKey, user, defaultValue)
}

// GetJSONVariation is a helper function to get a JSON variation from context.
func GetJSONVariation(ctx context.Context, flagKey string, defaultValue map[string]any) map[string]any {
	client, ok := GetClient(ctx)
	if !ok {
		return defaultValue
	}

	user, ok := GetUser(ctx)
	if !ok {
		user = NewAnonymousUser()
	}

	return client.GetJSONVariation(ctx, flagKey, user, defaultValue)
}
