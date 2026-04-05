// Package featureflag implements the FeatureFlag bounded context.
//
// Subdomain:   Generic
// Area:        admin
// Alternative: LaunchDarkly, Unleash, Flagsmith, PostHog feature flags
//
// Feature toggles with rule groups for gradual rollout and A/B testing.
// Generic capability used by essentially every SaaS — swap for a managed
// flag service when evaluation volume grows.
//
// See docs/architecture/context-map.md for the full strategic classification.
package featureflag
