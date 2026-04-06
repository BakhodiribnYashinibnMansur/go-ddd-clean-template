package http

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gct/internal/context/content/generic/file"
	"gct/internal/context/content/generic/file/application/command"
	"gct/internal/context/content/generic/file/application/query"
	"gct/internal/context/content/generic/file/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	miniogo "github.com/minio/minio-go/v7"
	_ "image/gif"
	_ "image/png"
)

// Handler provides HTTP endpoints for the File bounded context.
type Handler struct {
	bc     *file.BoundedContext
	l      logger.Log
	minio  *miniogo.Client
	bucket string
}

// NewHandler creates a new File HTTP handler.
func NewHandler(bc *file.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// SetMinio sets the MinIO client for upload/download operations.
func (h *Handler) SetMinio(client *miniogo.Client, bucket string) {
	h.minio = client
	h.bucket = bucket
}

// @Summary Create a file record
// @Description Create a new file record
// @Tags Files
// @Accept json
// @Produce json
// @Param request body CreateRequest true "File data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files [post]
// Create creates a new file record.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateFileCommand{
		Name:         req.Name,
		OriginalName: req.OriginalName,
		MimeType:     req.MimeType,
		Size:         req.Size,
		Path:         req.Path,
		URL:          req.URL,
		UploadedBy:   req.UploadedBy,
	}
	if err := h.bc.CreateFile.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List files
// @Description Get a paginated list of files
// @Tags Files
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files [get]
// List returns a paginated list of files.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListFilesQuery{
		Filter: domain.FileFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListFiles.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Files, "total": result.Total})
}

// @Summary Get a file
// @Description Get a file by ID
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/{id} [get]
// Get returns a single file by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseFileID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetFile.Handle(ctx.Request.Context(), query.GetFileQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Upload a single image
// @Description Upload a single image file, re-encoded as JPEG
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/upload/image [post]
// UploadImage handles POST /files/upload/image — single image upload, re-encoded as JPEG.
func (h *Handler) UploadImage(ctx *gin.Context) {
	if h.minio == nil {
		response.RespondWithError(ctx, httpx.ErrStorageNotConfigured, http.StatusServiceUnavailable)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrFileRequired, http.StatusBadRequest)
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrInvalidImage, http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	objectName := uuid.New().String() + ".jpeg"
	_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, &buf, int64(buf.Len()), miniogo.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": objectName})
}

// @Summary Upload multiple images
// @Description Upload multiple image files, re-encoded as JPEG
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param files[] formData file true "Files to upload"
// @Success 200 {object} map[string][]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/upload/images [post]
// UploadImages handles POST /files/upload/images — multiple image upload.
func (h *Handler) UploadImages(ctx *gin.Context) {
	if h.minio == nil {
		response.RespondWithError(ctx, httpx.ErrStorageNotConfigured, http.StatusServiceUnavailable)
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrMultipartRequired, http.StatusBadRequest)
		return
	}

	files := form.File["files"]
	var names []string
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			response.RespondWithError(ctx, err, http.StatusInternalServerError)
			return
		}

		img, _, err := image.Decode(src)
		src.Close()
		if err != nil {
			response.RespondWithError(ctx, apperrors.NewHandlerError(apperrors.ErrHandlerBadRequest, "invalid image: "+fh.Filename), http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})

		objectName := uuid.New().String() + ".jpeg"
		_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, &buf, int64(buf.Len()), miniogo.PutObjectOptions{ContentType: "image/jpeg"})
		if err != nil {
			response.RespondWithError(ctx, err, http.StatusInternalServerError)
			return
		}
		names = append(names, objectName)
	}

	ctx.JSON(http.StatusOK, gin.H{"data": names})
}

// @Summary Upload a document
// @Description Upload a single document file
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/upload/doc [post]
// UploadDoc handles POST /files/upload/doc — single document upload.
func (h *Handler) UploadDoc(ctx *gin.Context) {
	if h.minio == nil {
		response.RespondWithError(ctx, httpx.ErrStorageNotConfigured, http.StatusServiceUnavailable)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrFileRequired, http.StatusBadRequest)
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}
	defer src.Close()

	objectName := uuid.New().String() + filepath.Ext(fileHeader.Filename)
	_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, src, fileHeader.Size, miniogo.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")})
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": objectName})
}

// @Summary Download a file
// @Description Download a file by file path
// @Tags Files
// @Produce octet-stream
// @Param file-path query string true "File path"
// @Success 200 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/download [get]
// Download handles GET /files/download?file-path=...
func (h *Handler) Download(ctx *gin.Context) {
	filePath := ctx.Query("file-path")
	if filePath == "" {
		response.RespondWithError(ctx, httpx.ErrFilePathRequired, http.StatusBadRequest)
		return
	}

	// Serve local file
	if _, err := os.Stat(filePath); err == nil {
		ctx.File(filePath)
		return
	}

	// Try MinIO
	if h.minio != nil {
		obj, err := h.minio.GetObject(ctx.Request.Context(), h.bucket, filePath, miniogo.GetObjectOptions{})
		if err != nil {
			response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
			return
		}
		defer obj.Close()

		info, err := obj.Stat()
		if err != nil {
			response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
			return
		}

		ctx.DataFromReader(http.StatusOK, info.Size, info.ContentType, obj, nil)
		return
	}

	response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
}

// ensure imports are used
var _ = io.Discard
