package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain/mock"
	"github.com/gin-gonic/gin"
)

// UploadDoc handles document upload
func (h *Controller) UploadDoc(ctx *gin.Context) {
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return mock.FileInfoDocument().FileName }) {
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	fileContentType := file.Header[headerContentType][0]
	if fileContentType != docContentType && fileContentType != docxContentType && fileContentType != pdfContentType {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid file format", nil, false)
		return
	}

	fileMultipart, err := file.Open()
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	defer fileMultipart.Close()

	var docFileName string
	if fileContentType != pdfContentType {
		docFileName, err = h.useCase.Minio.UploadDoc(ctx.Request.Context(), fileMultipart, file.Size, fileContentType)
	} else {
		docFileName, err = h.useCase.Minio.UploadPDF(ctx.Request.Context(), fileMultipart, file.Size, fileContentType)
	}

	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, docFileName, nil, true)
}
