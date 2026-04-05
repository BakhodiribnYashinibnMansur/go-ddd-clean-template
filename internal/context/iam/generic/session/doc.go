// Package session implements the Session bounded context.
//
// Subdomain:   Generic
// Area:        iam
// Alternative: Keycloak sessions, Redis-backed session store
//
// Tracks active authenticated sessions, supports listing, single revocation,
// and revoke-all. Uses the User BC's SignOut commands via an adapter to
// preserve BC isolation.
//
// See docs/architecture/context-map.md for the full strategic classification.
package session
