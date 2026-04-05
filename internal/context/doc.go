// Package contexts groups all bounded contexts (Eric Evans DDD) of the system.
//
// # Folder Organization: Area → Tier → BC (Hybrid)
//
// Bounded contexts are grouped first by domain area (iam/, ops/, content/,
// admin/), then by Evans' strategic subdomain tier (generic/, supporting/,
// and core/ when present). For example: iam/generic/user, iam/supporting/audit.
//
// This hybrid layout preserves domain cohesion (all iam BCs live under iam/)
// while making the strategic tier visible in the import path. A BC's tier may
// change over time — when it does, move the BC between tier sub-folders within
// its area.
//
// Every area has core/, generic/, and supporting/ sub-folders. core/ folders
// may be empty (tracked via .gitkeep) — this marks the slot for a future Core
// BC without forcing one to exist prematurely.
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
