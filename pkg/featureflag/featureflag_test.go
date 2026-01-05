package featureflag_test

import (
	"context"
	"testing"

	"gct/pkg/featureflag"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	t.Run("NewUser creates user with key", func(t *testing.T) {
		user := featureflag.NewUser("user-123")
		assert.Equal(t, "user-123", user.Key)
		assert.False(t, user.Anonymous)
		assert.NotNil(t, user.Custom)
	})

	t.Run("NewAnonymousUser creates anonymous user", func(t *testing.T) {
		user := featureflag.NewAnonymousUser()
		assert.Equal(t, "anonymous", user.Key)
		assert.True(t, user.Anonymous)
	})

	t.Run("WithEmail sets email", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithEmail("test@example.com")
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("WithName sets name", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithName("John Doe")
		assert.Equal(t, "John Doe", user.Name)
	})

	t.Run("WithCountry sets country", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithCountry("US")
		assert.Equal(t, "US", user.Country)
	})

	t.Run("WithCustom adds custom attributes", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithCustom("plan", "premium").
			WithCustom("beta", true)

		assert.Equal(t, "premium", user.Custom["plan"])
		assert.Equal(t, true, user.Custom["beta"])
	})

	t.Run("Fluent builder pattern", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithEmail("test@example.com").
			WithName("John Doe").
			WithCountry("US").
			WithCustom("plan", "premium").
			WithCustom("beta", true)

		assert.Equal(t, "user-123", user.Key)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "US", user.Country)
		assert.Equal(t, "premium", user.Custom["plan"])
		assert.Equal(t, true, user.Custom["beta"])
	})

	t.Run("ToEvaluationContext converts user", func(t *testing.T) {
		user := featureflag.NewUser("user-123").
			WithEmail("test@example.com").
			WithName("John Doe")

		evalCtx := user.ToEvaluationContext()
		assert.NotNil(t, evalCtx)
	})
}

func TestContext(t *testing.T) {
	t.Run("WithUser adds user to context", func(t *testing.T) {
		ctx := context.Background()
		user := featureflag.NewUser("user-123")

		ctx = featureflag.WithUser(ctx, user)

		retrievedUser, ok := featureflag.GetUser(ctx)
		assert.True(t, ok)
		assert.Equal(t, user.Key, retrievedUser.Key)
	})

	t.Run("GetUser returns false when no user in context", func(t *testing.T) {
		ctx := context.Background()

		_, ok := featureflag.GetUser(ctx)
		assert.False(t, ok)
	})
}

// Example test demonstrating how to test code that uses feature flags
func TestFeatureFlagUsage(t *testing.T) {
	t.Run("Feature enabled for premium users", func(t *testing.T) {
		// This is a conceptual test - in real scenarios, you would:
		// 1. Initialize a test feature flag client with test configuration
		// 2. Create a user with specific attributes
		// 3. Test that the feature behaves correctly based on flag state

		user := featureflag.NewUser("test-user").
			WithCustom("plan", "premium")

		// In a real test, you would check the feature behavior
		assert.Equal(t, "premium", user.Custom["plan"])
	})

	t.Run("Feature disabled for free users", func(t *testing.T) {
		user := featureflag.NewUser("test-user").
			WithCustom("plan", "free")

		assert.Equal(t, "free", user.Custom["plan"])
	})
}

// BenchmarkUserCreation benchmarks user creation
func BenchmarkUserCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = featureflag.NewUser("user-123").
			WithEmail("test@example.com").
			WithName("John Doe").
			WithCountry("US").
			WithCustom("plan", "premium")
	}
}

// BenchmarkToEvaluationContext benchmarks context conversion
func BenchmarkToEvaluationContext(b *testing.B) {
	user := featureflag.NewUser("user-123").
		WithEmail("test@example.com").
		WithName("John Doe").
		WithCountry("US").
		WithCustom("plan", "premium")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.ToEvaluationContext()
	}
}
