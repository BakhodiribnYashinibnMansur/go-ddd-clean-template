// Package announcement implements the Announcement bounded context.
//
// Subdomain:      Supporting
// Area:           content
// Responsibility: Product-specific broadcast messaging to users.
//
// Supporting (not Generic) because scheduling, audience targeting, and
// localization rules are defined by the product. Distinct from the generic
// Notification BC — announcements are broadcast content with product semantics.
//
// See docs/architecture/context-map.md for the full strategic classification.
package announcement
