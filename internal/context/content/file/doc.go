// Package file implements the File bounded context.
//
// Subdomain:   Generic
// Area:        content
// Alternative: S3 SDK directly, UploadThing, Uploadcare, Cloudinary
//
// File upload/download/metadata on top of MinIO (S3-compatible). No
// product-specific logic beyond standard file handling.
//
// See docs/architecture/context-map.md for the full strategic classification.
package file
