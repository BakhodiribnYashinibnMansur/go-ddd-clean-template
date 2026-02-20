package admin

import (
	"net/http"
	"strconv"

	"gct/internal/domain"
	errorcoderepo "gct/internal/repo/persistent/postgres/errorcode"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ErrorCodes(ctx *gin.Context) {
	codes, err := h.uc.ErrorCode.List(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to fetch error codes", "error", err)
	}

	// Client-side filtering
	category := ctx.Query("category")
	severity := ctx.Query("severity")

	var filtered []*domain.ErrorCode
	for _, c := range codes {
		if category != "" && string(c.Category) != category {
			continue
		}
		if severity != "" && string(c.Severity) != severity {
			continue
		}
		filtered = append(filtered, c)
	}

	h.servePage(ctx, "error_codes.html", "Error Codes", "error_codes", map[string]any{
		"Codes":          filtered,
		"TotalCount":     len(codes),
		"FilterCategory": category,
		"FilterSeverity": severity,
	})
}

func (h *Handler) CreateErrorCodePost(ctx *gin.Context) {
	var req struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		HTTPStatus int    `json:"http_status"`
		Category   string `json:"category"`
		Severity   string `json:"severity"`
		Retryable  bool   `json:"retryable"`
		RetryAfter int    `json:"retry_after"`
		Suggestion string `json:"suggestion"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	input := errorcoderepo.CreateErrorCodeInput{
		Code:       req.Code,
		Message:    req.Message,
		HTTPStatus: req.HTTPStatus,
		Category:   domain.ErrorCategory(req.Category),
		Severity:   domain.ErrorSeverity(req.Severity),
		Retryable:  req.Retryable,
		RetryAfter: req.RetryAfter,
		Suggestion: req.Suggestion,
	}

	_, err := h.uc.ErrorCode.Create(ctx.Request.Context(), input)
	if err != nil {
		h.l.Errorw("failed to create error code", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Error code created"})
}

func (h *Handler) UpdateErrorCodePost(ctx *gin.Context) {
	code := ctx.Param("code")

	var req struct {
		Message    string `json:"message"`
		HTTPStatus int    `json:"http_status"`
		Category   string `json:"category"`
		Severity   string `json:"severity"`
		Retryable  bool   `json:"retryable"`
		RetryAfter int    `json:"retry_after"`
		Suggestion string `json:"suggestion"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	cat := domain.ErrorCategory(req.Category)
	sev := domain.ErrorSeverity(req.Severity)
	httpStatus := req.HTTPStatus

	input := errorcoderepo.UpdateErrorCodeInput{
		Message:    &req.Message,
		HTTPStatus: &httpStatus,
		Category:   &cat,
		Severity:   &sev,
		Retryable:  &req.Retryable,
		RetryAfter: &req.RetryAfter,
		Suggestion: &req.Suggestion,
	}

	_, err := h.uc.ErrorCode.Update(ctx.Request.Context(), code, input)
	if err != nil {
		h.l.Errorw("failed to update error code", "code", code, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Error code updated"})
}

func (h *Handler) DeleteErrorCodePost(ctx *gin.Context) {
	code := ctx.Param("code")

	err := h.uc.ErrorCode.Delete(ctx.Request.Context(), code)
	if err != nil {
		h.l.Errorw("failed to delete error code", "code", code, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Error code deleted"})
}

// Helper for error code template
func httpStatusText(code int) string {
	return strconv.Itoa(code) + " " + http.StatusText(code)
}
