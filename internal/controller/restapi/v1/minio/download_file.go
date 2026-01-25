package minio

import (
	"net/http"
	"os"

	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// DownloadFile godoc
// @Summary     Download file
// @Description Download a file by path
// @Tags        files
// @Produce     octet-stream
// @Param       file-path query string true "File path"
// @Success     200 {string} string "File content"
// @Failure     400 {object} response.ErrorResponse
// @Router      /files/download [get]
func (h *Controller) DownloadFile(ctx *gin.Context) {
	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return string("file_content_mock") }) {
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
