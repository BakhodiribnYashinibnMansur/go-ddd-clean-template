// Package systemerror implements the SystemError bounded context.
//
// Subdomain:   Generic
// Area:        ops
// Alternative: Sentry, Rollbar, Bugsnag, Honeybadger
//
// Captures runtime/system errors and tracks resolution workflow. Direct
// analogue to any error-tracking SaaS — keep API stable and internals simple.
//
// See docs/architecture/context-map.md for the full strategic classification.
package systemerror
