// Package ratelimit implements the RateLimit bounded context.
//
// Subdomain:   Generic
// Area:        ops
// Alternative: Nginx limit_req, Envoy rate limiter, Redis-based limiter
//
// Per-endpoint and per-client rate limiting configuration. Could live entirely
// at the reverse proxy layer in production; kept in-app for template
// simplicity and per-route customization.
//
// See docs/architecture/context-map.md for the full strategic classification.
package ratelimit
