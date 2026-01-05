package response

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gct/consts"
	apperrors "gct/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================================
// Error Response Structure
// ============================================================================

type ErrorDetail struct {
	Code       string `example:"RESOURCE_NOT_FOUND"                                          json:"code"`
	Message    string `example:"The requested resource was not found."                       json:"message"`
	Details    string `example:"The user with the ID '12345' does not exist in our records." json:"details,omitempty"`
	Timestamp  string `example:"2023-12-08T12:30:45Z"                                        json:"timestamp"`
	Path       string `example:"/api/v1/users/12345"                                         json:"path"`
	Method     string `example:"GET"                                                         json:"method"`
	Suggestion string `example:"Please check our documentation."                             json:"suggestion,omitempty"`

	// Enhanced fields
	Severity   string `example:"MEDIUM"                                                      json:"severity,omitempty"`
	Category   string `example:"DATA"                                                        json:"category,omitempty"`
	Retryable  bool   `example:"false"                                                       json:"retryable,omitempty"`
	RetryAfter int    `example:"5"                                                           json:"retry_after,omitempty"` // seconds
}

// Error for backward compatibility
type Error struct {
	Error string `example:"message" json:"error"`
}

type ErrorResponse struct {
	Status           string      `example:"error"                                                    json:"status"`
	StatusCode       int         `example:"404"                                                      json:"statusCode"`
	Error            ErrorDetail `json:"error"`
	RequestId        string      `example:"a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8"                     json:"requestId"`
	DocumentationUrl string      `example:"https://developer.mozilla.org/en-US/docs/Web/HTTP/Status" json:"documentation_url"`
}

// ============================================================================
// Constants
// ============================================================================

const (
	// DefaultDocsURL points to MDN HTTP Status codes reference
	DefaultDocsURL = "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status"
)

// statusSuggestions maps HTTP status codes to helpful suggestions based on MDN docs
var statusSuggestions = map[int]string{
	// 4xx Client Errors
	400: "The server cannot process the request due to a client error (e.g., malformed request syntax, invalid parameters).",
	401: "The request lacks valid authentication credentials. Please ensure you are logged in and have a valid token.",
	402: "Payment Required. This code is reserved for future use.",
	403: "The client does not have access rights to the content. Please check your permissions.",
	404: "The requested resource could not be found. Please check the URL ID or path.",
	405: "The request method is known by the server but is not supported by the target resource.",
	406: "The web server doesn't find any content that conforms to the criteria given by the user agent.",
	407: "Authentication is needed to be done by a proxy.",
	408: "The server timed out waiting for the request. Please try again.",
	409: "The request leaves the server in a conflict state (e.g., resource already exists).",
	410: "The requested content has been permanently deleted from server.",
	411: "The server rejected the request because the Content-Length header field is not defined.",
	412: "The client has indicated preconditions in its headers which the server does not meet.",
	413: "The request body is larger than limits defined by server.",
	414: "The URI requested by the client is longer than the server is willing to interpret.",
	415: "The media format of the requested data is not supported by the server.",
	416: "The ranges specified by the Range header field in the request cannot be fulfilled.",
	417: "The expectation indicated by the Expect request header field cannot be met by the server.",
	418: "The server refuses the attempt to brew coffee with a teapot.",
	421: "The request was directed at a server that is not able to produce a response.",
	422: "The request was well-formed but was unable to be followed due to semantic errors (validation failed).",
	423: "The resource that is being accessed is locked.",
	424: "The request failed due to failure of a previous request.",
	425: "Indicates that the server is unwilling to risk processing a request that might be replayed.",
	426: "The server refuses to perform the request using the current protocol.",
	428: "The origin server requires the request to be conditional.",
	429: "You have sent too many requests in a given amount of time. Please retry after some time.",
	431: "The server is unwilling to process the request because its header fields are too large.",
	451: "The user agent requested a resource that cannot legally be provided.",
	499: "Client Closed Request. The client closed the connection before the server could send a response.",

	// 5xx Server Errors
	500: "The server has encountered a situation it does not know how to handle. Please contact support.",
	501: "The request method is not supported by the server and cannot be handled.",
	502: "The server received an invalid response from the upstream server.",
	503: "The server is not ready to handle the request (maintenance or overloaded). Please try again later.",
	504: "The server is acting as a gateway and cannot get a response in time.",
	505: "The HTTP version used in the request is not supported by the server.",
	506: "The server has an internal configuration error regarding content negotiation.",
	507: "The server is unable to store the representation needed to successfully complete the request.",
	508: "The server detected an infinite loop while processing the request.",
	510: "Further extensions to the request are required for the server to fulfill it.",
	511: "The client needs to authenticate to gain network access.",
}

// ============================================================================
// Response Helpers
// ============================================================================

// RespondWithError sends the error response in the defined JSON format
func RespondWithError(c *gin.Context, err error, fallbackCode int) {
	_ = c.Error(err)
	// 1. Parse error to get status code and details
	status, errResp := parseErrorToResponse(c, err, fallbackCode)

	// 2. Send JSON response
	c.JSON(status, errResp)
}

// parseErrorToResponse converts error to ErrorResponse structure
func parseErrorToResponse(c *gin.Context, err error, fallbackCode int) (int, ErrorResponse) {
	var (
		statusCode = 500
		errorCode  = apperrors.ErrInternal
		message    = "An unexpected error occurred."
		details    = ""
		severity   = ""
		category   = ""
		retryable  = false
		retryAfter = 0
		// fields     map[string]any // unused
	)

	// Get language from Accept-Language header
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = "en" // default to English
	}
	// Extract primary language (e.g., "en-US" -> "en", "uz-UZ" -> "uz")
	if len(lang) >= 2 {
		lang = lang[:2]
	}

	// Check if it's our AppError
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		statusCode = MapToHTTPStatus(appErr.Code)

		errorCode = appErr.Type
		if errorCode == "" {
			errorCode = appErr.Code
		}

		// Get user-friendly message in requested language
		message = apperrors.GetUserMessage(errorCode, lang)
		if message == "" {
			message = appErr.Message
		}

		details = appErr.Details

		// Get metadata
		meta := appErr.GetMetadata()
		severity = string(meta.Severity)
		category = string(meta.Category)
		retryable = meta.Retryable

		// Calculate retry after based on strategy
		if retryable {
			switch meta.RetryStrategy {
			case apperrors.RetryStrategyImmediate:
				retryAfter = 1
			case apperrors.RetryStrategyLinear:
				retryAfter = 5
			case apperrors.RetryStrategyExponential:
				retryAfter = 10
			}
		}

		// fields = appErr.Fields // unused
	} else if err != nil {
		message = err.Error()
		if fallbackCode != 0 {
			statusCode = fallbackCode
		}
	}

	// Request ID
	reqID := c.GetHeader(consts.HeaderXRequestID)
	if reqID == "" {
		reqID = uuid.New().String()
	}

	// Dynamic Suggestion based on Status Code
	suggestion, exists := statusSuggestions[statusCode]
	if !exists {
		if statusCode >= 500 {
			suggestion = statusSuggestions[500]
		} else if statusCode >= 400 {
			suggestion = "Please check your request and try again."
		} else {
			suggestion = "Operation completed with status: " + strconv.Itoa(statusCode)
		}
	}

	// Build Documentation URL
	// We can append the status code to link directly to MDN standard docs
	// Example: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/404
	docURL := fmt.Sprintf("%s/%d", DefaultDocsURL, statusCode)

	return statusCode, ErrorResponse{
		Status:     consts.ResponseStatusError,
		StatusCode: statusCode,
		Error: ErrorDetail{
			Code:       errorCode,
			Message:    message,
			Details:    details,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Path:       c.Request.URL.Path,
			Method:     c.Request.Method,
			Suggestion: suggestion,
			Severity:   severity,
			Category:   category,
			Retryable:  retryable,
			RetryAfter: retryAfter,
		},
		RequestId:        reqID,
		DocumentationUrl: docURL,
	}
}

// MapToHTTPStatus maps error code to HTTP status code
func MapToHTTPStatus(code string) int {
	if status := mapRepoStatus(code); status != 0 {
		return status
	}
	if status := mapServiceStatus(code); status != 0 {
		return status
	}
	if status := mapHandlerStatus(code); status != 0 {
		return status
	}
	return 500
}

func mapRepoStatus(code string) int {
	switch code {
	case apperrors.CodeRepoNotFound, apperrors.ErrRepoNotFound:
		return 404
	case apperrors.CodeRepoAlreadyExists, apperrors.ErrRepoAlreadyExists:
		return 409
	case apperrors.CodeRepoTimeout, apperrors.ErrRepoTimeout:
		return 504
	case apperrors.CodeUserNotFound, apperrors.ErrUserNotFound,
		apperrors.CodeSessionNotFound, apperrors.ErrSessionNotFound:
		return 404
	default:
		return 0
	}
}

func mapServiceStatus(code string) int {
	switch code {
	case apperrors.CodeServiceInvalidInput, apperrors.ErrServiceInvalidInput,
		apperrors.CodeServiceValidation, apperrors.ErrServiceValidation,
		apperrors.ErrBadRequest, apperrors.CodeBadRequest,
		apperrors.ErrInvalidInput, apperrors.CodeInvalidInput,
		apperrors.ErrValidation, apperrors.CodeValidation:
		return 400
	case apperrors.CodeServiceNotFound, apperrors.ErrServiceNotFound,
		apperrors.ErrNotFound, apperrors.CodeNotFound:
		return 404
	case apperrors.CodeServiceAlreadyExists, apperrors.ErrServiceAlreadyExists,
		apperrors.CodeServiceConflict, apperrors.ErrServiceConflict:
		return 409
	case apperrors.CodeServiceUnauthorized, apperrors.ErrServiceUnauthorized:
		return 401
	case apperrors.CodeServiceForbidden, apperrors.ErrServiceForbidden:
		return 403
	case apperrors.CodeServiceBusinessRule, apperrors.ErrServiceBusinessRule:
		return 422
	case apperrors.CodeServiceDependency, apperrors.ErrServiceDependency:
		return 502
	default:
		return 0
	}
}

func mapHandlerStatus(code string) int {
	switch code {
	case "HANDLER_BAD_REQUEST", "4000":
		return 400
	case "HANDLER_UNAUTHORIZED", "4001":
		return 401
	case "HANDLER_FORBIDDEN", "4003":
		return 403
	case "HANDLER_NOT_FOUND", "4004":
		return 404
	case "HANDLER_CONFLICT", "4009":
		return 409
	default:
		return 0
	}
}
