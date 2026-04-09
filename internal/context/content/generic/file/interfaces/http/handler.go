package http

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"fmt"
	"strconv"
	"strings"

	"gct/internal/context/content/generic/file"
	"gct/internal/context/content/generic/file/application/command"
	"gct/internal/context/content/generic/file/application/query"
	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/transcode"

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
		Filter: filerepo.FileFilter{Limit: pg.Limit, Offset: pg.Offset},
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
	id, err := fileentity.ParseFileID(ctx.Param("id"))
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

// allowedVideoMIME lists MIME types accepted by UploadVideo.
var allowedVideoMIME = map[string]bool{
	"video/mp4":          true,
	"video/webm":         true,
	"video/quicktime":    true,
	"video/x-matroska":   true,
	"video/avi":          true,
	"video/x-msvideo":    true,
	"video/x-flv":        true,
	"video/3gpp":         true,
	"application/octet-stream": true, // browsers sometimes send this
}

// @Summary Upload a video
// @Description Upload a video file, transcoded to MP4 (H.264 + AAC) with progressive streaming support
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Video file to upload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/upload/video [post]
// UploadVideo handles POST /files/upload/video — video upload with MP4 transcode.
func (h *Handler) UploadVideo(ctx *gin.Context) {
	if h.minio == nil {
		response.RespondWithError(ctx, httpx.ErrStorageNotConfigured, http.StatusServiceUnavailable)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrFileRequired, http.StatusBadRequest)
		return
	}

	ct := fileHeader.Header.Get("Content-Type")
	if !allowedVideoMIME[ct] {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		videoExts := map[string]bool{".mp4": true, ".webm": true, ".mov": true, ".mkv": true, ".avi": true, ".flv": true, ".3gp": true}
		if !videoExts[ext] {
			response.RespondWithError(ctx, httpx.ErrUnsupportedVideo, http.StatusBadRequest)
			return
		}
	}

	// Save uploaded file to temp directory.
	tmpDir, err := os.MkdirTemp("", "video-upload-*")
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input"+filepath.Ext(fileHeader.Filename))
	if err := ctx.SaveUploadedFile(fileHeader, inputPath); err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	// Transcode to MP4.
	objectName := uuid.New().String() + ".mp4"
	outputPath := filepath.Join(tmpDir, objectName)

	if err := transcode.ToMP4(ctx.Request.Context(), inputPath, outputPath); err != nil {
		h.l.Errorc(ctx.Request.Context(), "video transcode failed", "error", err)
		response.RespondWithError(ctx, httpx.ErrTranscodeFailed, http.StatusInternalServerError)
		return
	}

	// Upload transcoded MP4 to MinIO.
	outFile, err := os.Open(outputPath)
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	stat, err := outFile.Stat()
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	_, err = h.minio.PutObject(ctx.Request.Context(), h.bucket, objectName, outFile, stat.Size(), miniogo.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		response.RespondWithError(ctx, err, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": objectName})
}

// @Summary Stream a file
// @Description Stream a file with HTTP Range Request support (progressive MP4 streaming)
// @Tags Files
// @Produce octet-stream
// @Param file-path query string true "File path in storage"
// @Success 200 {file} binary
// @Success 206 {file} binary
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /files/stream [get]
// Stream handles GET /files/stream?file-path=... — progressive streaming with range requests.
func (h *Handler) Stream(ctx *gin.Context) {
	filePath := ctx.Query("file-path")
	if filePath == "" {
		response.RespondWithError(ctx, httpx.ErrFilePathRequired, http.StatusBadRequest)
		return
	}

	if h.minio == nil {
		response.RespondWithError(ctx, httpx.ErrStorageNotConfigured, http.StatusServiceUnavailable)
		return
	}

	// Get file info for total size and content type.
	info, err := h.minio.StatObject(ctx.Request.Context(), h.bucket, filePath, miniogo.StatObjectOptions{})
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
		return
	}

	totalSize := info.Size
	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	rangeHeader := ctx.GetHeader("Range")
	if rangeHeader == "" {
		// No range — serve full file.
		obj, err := h.minio.GetObject(ctx.Request.Context(), h.bucket, filePath, miniogo.GetObjectOptions{})
		if err != nil {
			response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
			return
		}
		defer obj.Close()

		ctx.Header("Accept-Ranges", "bytes")
		ctx.Header("Content-Length", strconv.FormatInt(totalSize, 10))
		ctx.DataFromReader(http.StatusOK, totalSize, contentType, obj, nil)
		return
	}

	// Parse Range header: "bytes=start-end"
	start, end, ok := parseRange(rangeHeader, totalSize)
	if !ok {
		ctx.Header("Content-Range", fmt.Sprintf("bytes */%d", totalSize))
		ctx.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	length := end - start + 1

	opts := miniogo.GetObjectOptions{}
	opts.SetRange(start, end)

	obj, err := h.minio.GetObject(ctx.Request.Context(), h.bucket, filePath, opts)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrFileNotFound, http.StatusNotFound)
		return
	}
	defer obj.Close()

	ctx.Header("Accept-Ranges", "bytes")
	ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize))
	ctx.Header("Content-Length", strconv.FormatInt(length, 10))
	ctx.DataFromReader(http.StatusPartialContent, length, contentType, obj, nil)
}

// parseRange parses an HTTP Range header value like "bytes=0-1023" and returns
// the start and end byte positions. Returns ok=false for malformed ranges.
func parseRange(rangeHeader string, totalSize int64) (start, end int64, ok bool) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return 0, 0, false
	}

	spec := strings.TrimPrefix(rangeHeader, prefix)
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	if parts[0] == "" {
		// Suffix range: "bytes=-500" means last 500 bytes.
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 {
			return 0, 0, false
		}
		start = totalSize - suffix
		if start < 0 {
			start = 0
		}
		return start, totalSize - 1, true
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= totalSize {
		return 0, 0, false
	}

	if parts[1] == "" {
		// Open-ended: "bytes=500-"
		return start, totalSize - 1, true
	}

	end, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || end < start {
		return 0, 0, false
	}
	if end >= totalSize {
		end = totalSize - 1
	}

	return start, end, true
}

// ensure imports are used
var _ = io.Discard
