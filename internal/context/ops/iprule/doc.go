// Package iprule implements the IPRule bounded context.
//
// Subdomain:      Supporting
// Area:           ops
// Responsibility: Allow/deny IP access rules encoding business risk appetite.
//
// Supporting (not Generic) because the rule semantics (who can access what
// from which network) are a business-specific security policy, not generic
// firewalling.
//
// See docs/architecture/context-map.md for the full strategic classification.
package iprule
