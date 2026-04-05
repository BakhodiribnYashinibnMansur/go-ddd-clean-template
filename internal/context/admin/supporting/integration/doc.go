// Package integration implements the Integration bounded context.
//
// Subdomain:      Supporting
// Area:           admin
// Responsibility: Registry of outbound third-party integrations and their config.
//
// Supporting (not Generic) because which third parties are integrated and
// how they are configured is product-specific knowledge. The registry pattern
// is generic; the contents are not.
//
// See docs/architecture/context-map.md for the full strategic classification.
package integration
