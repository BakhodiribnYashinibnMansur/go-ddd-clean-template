// Package notification implements the Notification bounded context.
//
// Subdomain:   Generic
// Area:        content
// Alternative: SendGrid, Twilio, Novu, Postmark
//
// Generic multi-channel notification dispatch (email, SMS, push, in-app).
// Any product needs this; the implementation here is deliberately simple and
// swap-ready.
//
// See docs/architecture/context-map.md for the full strategic classification.
package notification
