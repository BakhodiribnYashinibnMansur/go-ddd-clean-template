// Package sitesetting implements the SiteSetting bounded context.
//
// Subdomain:      Supporting
// Area:           admin
// Responsibility: Platform-wide configuration values and their semantics.
//
// Supporting (not Generic) because the setting keys, validation rules, and
// effects on runtime behavior are all product-defined. Generic CRUD mechanics
// with product-specific meaning.
//
// See docs/architecture/context-map.md for the full strategic classification.
package sitesetting
