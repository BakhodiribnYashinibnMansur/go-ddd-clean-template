// Package audit implements the Audit bounded context.
//
// Subdomain:      Supporting
// Area:           iam
// Responsibility: Compliance-grade activity trail (GDPR, SOC2, internal policy).
//
// Supporting (not Generic) because audit semantics are business-specific:
// what to log, retention periods, who can read logs, and how to correlate
// with endpoint history — all encode compliance posture of the product.
//
// See docs/architecture/context-map.md for the full strategic classification.
package audit
