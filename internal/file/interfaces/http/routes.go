package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all File HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/files")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("/upload/image", h.UploadImage)
	g.POST("/upload/images", h.UploadImages)
	g.POST("/upload/doc", h.UploadDoc)
	g.GET("/download", h.Download)
}
