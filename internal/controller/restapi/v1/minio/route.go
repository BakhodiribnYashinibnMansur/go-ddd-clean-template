package minio

import (
	"github.com/gin-gonic/gin"
)

func MinioRoute(h *gin.RouterGroup, minio *Controller, authMiddleware gin.HandlerFunc) {
	upload := h.Group("/v1/upload", authMiddleware)
	{
		upload.POST("/images", minio.UploadImages)
		upload.POST("/image", minio.UploadImage)
		upload.POST("/doc", minio.UploadDoc)
		upload.POST("/video", minio.UploadVideo)
	}
	download := h.Group("/v1/download")
	{
		download.GET("", minio.DownloadFile)
	}
	transfer := h.Group("/v1/transfer", authMiddleware)
	{
		transfer.POST("", minio.TransferFile)
	}
}
