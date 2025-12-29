package minio

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
)

// UploadImage handles single image upload
func (h *Controller) UploadImage(ctx *gin.Context) {
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

	imageFileName, err := h.useCase.Minio.UploadImage(fileMultipart, file.Size, imageContentType)
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, imageFileName, nil, true)
}

// UploadImages handles multiple image uploads
func (h *Controller) UploadImages(ctx *gin.Context) {
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

		imageFileName, err := h.useCase.Minio.UploadImage(fileMultipart, file.Size, imageContentType)
		fileMultipart.Close() // Close immediately

		if err != nil {
			response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
			return
		}
		uploadedFiles = append(uploadedFiles, imageFileName)
	}
	response.ControllerResponse(ctx, http.StatusOK, uploadedFiles, nil, true)
}
