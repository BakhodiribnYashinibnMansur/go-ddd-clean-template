package util

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func FileTransfer(ctx *gin.Context, filePath, contentType string) (err error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename="+path.Base(filePath))
	ctx.Data(http.StatusOK, "application/octet-stream", bytes)
	ctx.Writer.Header().Set("Content-Type", contentType)
	return nil
}

func DownloadFile(ctx *gin.Context, filePath string) error {
	_, fileName := filepath.Split(filePath)
	bytes, err := os.ReadFile("./" + filePath)
	if err != nil {
		return err
	}
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	if _, err := ctx.Writer.Write(bytes); err != nil {
		return err
	}
	return nil
}
