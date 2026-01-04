package minio

import (
	"net/http"
	"os"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"github.com/gin-gonic/gin"
)

// DownloadFile handles file download
func (h *Controller) DownloadFile(ctx *gin.Context) {
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return string("file_content_mock") }) {
		return
	}
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
