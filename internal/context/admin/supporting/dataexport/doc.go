// Package dataexport implements the DataExport bounded context.
//
// Subdomain:      Supporting
// Area:           admin
// Responsibility: GDPR "right to data portability" data extraction requests.
//
// Supporting (not Generic) because jurisdiction- and compliance-specific
// rules govern what data is exportable, in what format, and how the request
// is fulfilled.
//
// See docs/architecture/context-map.md for the full strategic classification.
package dataexport
