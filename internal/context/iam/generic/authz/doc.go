// Package authz implements the Authorization bounded context.
//
// Subdomain:   Generic
// Area:        iam
// Alternative: Casbin, OpenFGA, Keycloak RBAC, Oso
//
// Roles, permissions, policies, and scopes — a standard RBAC + policy
// evaluation engine. Strategic choice to keep in-house for template
// simplicity; swap for an authorization service as scale demands.
//
// See docs/architecture/context-map.md for the full strategic classification.
package authz
