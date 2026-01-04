package minio

import (
	"github.com/gin-gonic/gin"
)

func MinioRoute(h *gin.RouterGroup, minio *Controller, authMiddleware gin.HandlerFunc, csrfMiddleware gin.HandlerFunc) {
	upload := h.Group("files/upload")
	upload.Use(authMiddleware)
	upload.Use(csrfMiddleware)
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
	transfer := h.Group("files/transfer")
	transfer.Use(authMiddleware)
	transfer.Use(csrfMiddleware)
	{
		transfer.POST("", minio.TransferFile)
	}
}
