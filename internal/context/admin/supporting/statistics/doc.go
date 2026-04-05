// Package statistics implements the Statistics bounded context.
//
// Subdomain:      Supporting
// Area:           admin
// Responsibility: Business KPIs aggregated across other BCs.
//
// Supporting (not Generic) because the aggregations (user stats, content
// stats, error stats, session stats, integration stats, etc.) are defined
// per product. Read-only BC that queries other BCs' stores via ACL ports.
//
// See docs/architecture/context-map.md for the full strategic classification.
package statistics
