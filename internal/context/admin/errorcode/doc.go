// Package errorcode implements the ErrorCode bounded context.
//
// Subdomain:      Supporting
// Area:           admin
// Responsibility: Catalog of API error codes (public contract with clients).
//
// Supporting (not Generic) because the error codes, their messages, and their
// HTTP status mappings are part of the product's public API contract — not a
// generic mechanism.
//
// See docs/architecture/context-map.md for the full strategic classification.
package errorcode
