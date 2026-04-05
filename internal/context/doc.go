// Package contexts groups all bounded contexts (Eric Evans DDD) of the system.
//
// # Folder Organization: by Domain Area
//
// Bounded contexts are grouped by domain area (iam/, ops/, content/, admin/),
// NOT by strategic subdomain tier (core/supporting/generic). Strategic tier is
// a decision-making tool — documented in the Context Map, not enforced by
// folder names. A BC's tier can change over time without moving the code.
//
// # Strategic Classification
//
// Each BC is classified as Core, Supporting, or Generic per Evans' DDD
// strategic design patterns. See docs/architecture/context-map.md for the
// full classification and the reasoning behind each tier assignment.
//
// Each BC also declares its tier in its own doc.go via a "Subdomain:" marker
// comment, so the classification is visible at the package level.
//
// # Isolation Rules
//
// No bounded context may import another bounded context directly.
// Communication flows only through:
//   - gct/internal/contract/events — Published Language (domain events)
//   - gct/internal/contract/ports  — Anti-Corruption Layer interfaces
package contexts
