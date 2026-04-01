package http

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"gct/internal/file"
	"gct/internal/file/application/command"
	"gct/internal/file/application/query"
	"gct/internal/file/domain"
	"gct/internal/shared/infrastructure/logger"

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

// Create creates a new file record.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of files.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListFilesQuery{
		Filter: domain.FileFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListFiles.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Files, "total": result.Total})
}

// Get returns a single file by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetFile.Handle(ctx.Request.Context(), query.GetFileQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// UploadImage handles POST /files/upload/image — single image upload, re-encoded as JPEG.
func (h *Handler) UploadImage(ctx *gin.Context) {
	if h.minio == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	objectName := uuid.New().String() + ".jpeg"
	_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, &buf, int64(buf.Len()), miniogo.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": objectName})
}

// UploadImages handles POST /files/upload/images — multiple image upload.
func (h *Handler) UploadImages(ctx *gin.Context) {
	if h.minio == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}

	files := form.File["files"]
	var names []string
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		img, _, err := image.Decode(src)
		src.Close()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid image: " + fh.Filename})
			return
		}

		var buf bytes.Buffer
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})

		objectName := uuid.New().String() + ".jpeg"
		_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, &buf, int64(buf.Len()), miniogo.PutObjectOptions{ContentType: "image/jpeg"})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		names = append(names, objectName)
	}

	ctx.JSON(http.StatusOK, gin.H{"data": names})
}

// UploadDoc handles POST /files/upload/doc — single document upload.
func (h *Handler) UploadDoc(ctx *gin.Context) {
	if h.minio == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage not configured"})
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	objectName := uuid.New().String() + filepath.Ext(fileHeader.Filename)
	_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, src, fileHeader.Size, miniogo.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": objectName})
}

// Download handles GET /files/download?file-path=...
func (h *Handler) Download(ctx *gin.Context) {
	filePath := ctx.Query("file-path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file-path required"})
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
			ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		defer obj.Close()

		info, err := obj.Stat()
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		ctx.DataFromReader(http.StatusOK, info.Size, info.ContentType, obj, nil)
		return
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
}

// ensure imports are used
var _ = io.Discard
