// Package user implements the User bounded context.
//
// Subdomain:   Generic
// Area:        iam
// Alternative: Keycloak, Auth0, Firebase Auth, Ory Kratos
//
// Handles user accounts, sign-up/sign-in/sign-out, profile, and lifecycle
// (approve, role change, bulk actions). Custom in-house implementation kept
// intentionally simple; replace with an identity provider if the team scales.
//
// See docs/architecture/context-map.md for the full strategic classification.
package user
