// Package minio provides endpoints for file management using Minio/S3 compatible storage.
package minio

import (
	"github.com/gin-gonic/gin"
)

// MinioRoute registers routes for file uploads, downloads, and transfers.
// Authentication and CSRF protection are applied to mutation endpoints (upload and transfer).
func MinioRoute(h *gin.RouterGroup, minio *Controller, authMiddleware gin.HandlerFunc, authzMiddleware gin.HandlerFunc, csrfMiddleware gin.HandlerFunc) {
	// UPLOAD Group: Restricted to authenticated users with CSRF tokens.
	upload := h.Group("files/upload")
	upload.Use(authMiddleware)
	upload.Use(authzMiddleware)
	upload.Use(csrfMiddleware)
	{
		upload.POST("/images", minio.UploadImages) // Batch upload for image files.
		upload.POST("/image", minio.UploadImage)   // Single image file upload.
		upload.POST("/doc", minio.UploadDoc)       // Document/PDF file upload.
		upload.POST("/video", minio.UploadVideo)   // Video file upload.
	}

	// DOWNLOAD Group: Public access for file retrieval.
	download := h.Group("files/download")
	download.Use(authzMiddleware)
	{
		download.GET("", minio.DownloadFile)
	}

	// TRANSFER Group: Restricted endpoints for moving/managing files within buckets.
	transfer := h.Group("files/transfer")
	transfer.Use(authMiddleware)
	transfer.Use(authzMiddleware)
	transfer.Use(csrfMiddleware)
	{
		transfer.POST("", minio.TransferFile)
	}
	// FILE MANAGEMENT Group: CRUD on file_metadata table.
	files := h.Group("files")
	files.Use(authMiddleware)
	files.Use(authzMiddleware)
	{
		files.GET("/list", minio.ListFiles)
		files.PUT("/:id", csrfMiddleware, minio.UpdateFile)
		files.DELETE("/:id", csrfMiddleware, minio.DeleteFile)
	}
}
