package minio

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
)

// DownloadFile handles file download
func (h *Controller) DownloadFile(ctx *gin.Context) {
	pathStr := ctx.Query(filePath)
	if pathStr == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "file path required", nil, false)
		return
	}

	if _, err := os.Stat(pathStr); os.IsNotExist(err) {
		response.ControllerResponse(ctx, http.StatusBadRequest, "file not found", nil, false)
		return
	}

	ctx.File(pathStr)
}
