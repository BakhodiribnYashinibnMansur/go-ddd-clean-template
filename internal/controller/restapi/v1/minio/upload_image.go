package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"gct/internal/domain/mock"

	"github.com/gin-gonic/gin"
)

// UploadImage handles single image upload
func (h *Controller) UploadImage(ctx *gin.Context) {
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return mock.FileInfoImage().FileName }) {
		return
	}
	file, err := ctx.FormFile(formFileName)
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	imageContentType := file.Header[headerContentType][0]
	if !imageContentTypes[imageContentType] {
		response.ControllerResponse(ctx, http.StatusBadRequest, ErrInvalidFileFormat, nil, false)
		return
	}

	fileMultipart, err := file.Open()
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	defer fileMultipart.Close()

	imageFileName, err := h.useCase.Minio.UploadImage(ctx.Request.Context(), fileMultipart, file.Size, imageContentType)
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, imageFileName, nil, true)
}

// UploadImages handles multiple image uploads
func (h *Controller) UploadImages(ctx *gin.Context) {
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGets, func(count int) any {
		files := mock.FileInfos(count)
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.FileName
		}
		return names
	}) {
		return
	}
	var uploadedFiles []string
	form, err := ctx.MultipartForm()
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	files := form.File["files"]
	for _, file := range files {
		imageContentType := file.Header[headerContentType][0]
		if !imageContentTypes[imageContentType] {
			continue
		}

		fileMultipart, err := file.Open()
		if err != nil {
			response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
			return
		}

		imageFileName, err := h.useCase.Minio.UploadImage(ctx.Request.Context(), fileMultipart, file.Size, imageContentType)
		fileMultipart.Close() // Close immediately

		if err != nil {
			response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
			return
		}
		uploadedFiles = append(uploadedFiles, imageFileName)
	}
	response.ControllerResponse(ctx, http.StatusOK, uploadedFiles, nil, true)
}
