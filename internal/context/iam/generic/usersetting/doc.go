// Package usersetting implements the UserSetting bounded context.
//
// Subdomain: Generic
// Area:      iam
//
// Trivial per-user key/value preferences (upsert/list/delete). Intentionally
// minimal — no off-the-shelf SaaS equivalent because storing user preferences
// next to the user record is already generic CRUD.
//
// See docs/architecture/context-map.md for the full strategic classification.
package usersetting
