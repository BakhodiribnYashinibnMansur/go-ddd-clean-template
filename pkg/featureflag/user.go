package featureflag

import (
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// User represents a user for feature flag evaluation.
type User struct {
	Key       string                 // Unique identifier for the user
	Anonymous bool                   // Whether the user is anonymous
	Email     string                 // User email
	Name      string                 // User name
	Country   string                 // User country
	Custom    map[string]interface{} // Custom attributes
}

// NewUser creates a new user for feature flag evaluation.
func NewUser(key string) User {
	return User{
		Key:    key,
		Custom: make(map[string]interface{}),
	}
}

// NewAnonymousUser creates a new anonymous user.
func NewAnonymousUser() User {
	return User{
		Key:       "anonymous",
		Anonymous: true,
		Custom:    make(map[string]interface{}),
	}
}

// WithEmail sets the user email.
func (u User) WithEmail(email string) User {
	u.Email = email
	return u
}

// WithName sets the user name.
func (u User) WithName(name string) User {
	u.Name = name
	return u
}

// WithCountry sets the user country.
func (u User) WithCountry(country string) User {
	u.Country = country
	return u
}

// WithCustom adds a custom attribute.
func (u User) WithCustom(key string, value interface{}) User {
	if u.Custom == nil {
		u.Custom = make(map[string]interface{})
	}
	u.Custom[key] = value
	return u
}

// ToEvaluationContext converts the User to ffcontext.EvaluationContext.
func (u User) ToEvaluationContext() ffcontext.EvaluationContext {
	builder := ffcontext.NewEvaluationContextBuilder(u.Key)

	if u.Anonymous {
		builder.AddCustom("anonymous", true)
	}

	if u.Email != "" {
		builder.AddCustom("email", u.Email)
	}

	if u.Name != "" {
		builder.AddCustom("name", u.Name)
	}

	if u.Country != "" {
		builder.AddCustom("country", u.Country)
	}

	// Add all custom attributes
	for key, value := range u.Custom {
		builder.AddCustom(key, value)
	}

	return builder.Build()
}
