package httpx

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func FileTransfer(ctx *gin.Context, filePath, contentType string) (err error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	ctx.Header(HeaderContentDescription, FileTransferDescription)
	ctx.Header(HeaderContentDisposition, AttachmentPrefix+path.Base(filePath))
	ctx.Data(http.StatusOK, ContentTypeOctetStream, bytes)
	ctx.Writer.Header().Set(HeaderContentType, contentType)
	return nil
}

func DownloadFile(ctx *gin.Context, filePath string) error {
	_, fileName := filepath.Split(filePath)
	bytes, err := os.ReadFile(CurrentDir + filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	ctx.Header(HeaderContentType, ContentTypeOctetStream)
	ctx.Header(HeaderContentDisposition, AttachmentPrefix+fileName)
	if _, err := ctx.Writer.Write(bytes); err != nil {
		return fmt.Errorf("failed to write file response: %w", err)
	}
	return nil
}
