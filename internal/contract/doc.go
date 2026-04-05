// Package contracts holds cross-bounded-context stable integration surfaces.
//
//   - events/ — Published Language: versioned event DTOs consumed by subscribers.
//   - ports/  — consumer-defined interfaces (Anti-Corruption Layer) implemented
//     by producer bounded contexts and wired in the composition root.
//
// Contracts are the only way bounded contexts talk to each other.
package contracts
