// Package ratelimit provides authentication-specific rate limiting with
// per-IP sliding-window throttling and per-user account lockout with
// exponential backoff. State is stored in Redis.
//
// This package is independent of the admin-configurable API rate limiter
// located in internal/context/ops/generic/ratelimit/.
package ratelimit
