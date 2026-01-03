package minio

import (
	"github.com/gin-gonic/gin"
)

func MinioRoute(h *gin.RouterGroup, minio *Controller, authMiddleware gin.HandlerFunc) {
	upload := h.Group("files/upload", authMiddleware)
	{
		upload.POST("/images", minio.UploadImages)
		upload.POST("/image", minio.UploadImage)
		upload.POST("/doc", minio.UploadDoc)
		upload.POST("/video", minio.UploadVideo)
	}
	download := h.Group("files/download")
	{
		download.GET("", minio.DownloadFile)
	}
	transfer := h.Group("files/transfer", authMiddleware)
	{
		transfer.POST("", minio.TransferFile)
	}
}
