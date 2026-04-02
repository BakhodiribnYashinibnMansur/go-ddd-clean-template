package middleware

import (
	"fmt"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/contextx"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func getStatusColor(status int) (fg string, bg string) {
	switch {
	case status >= 200 && status < 300:
		return logger.ColorBlack, logger.BgBrightGreen // Success
	case status >= 300 && status < 400:
		return logger.ColorBlack, logger.BgBrightCyan // Redirect
	case status >= 400 && status < 500:
		return logger.ColorBlack, logger.BgBrightYellow // Client error
	case status >= 500:
		return logger.ColorBrightWhite, logger.BgRed // Server error
	default:
		return logger.ColorBrightWhite, logger.BgGray
	}
}

func getMethodStyle(method string) (fg string, bg string, label string) {
	switch method {
	case "GET":
		return logger.ColorBlack, logger.BgCyan, " GET "
	case "POST":
		return logger.ColorBlack, logger.BgGreen, " POST "
	case "PUT":
		return logger.ColorBlack, logger.BgOrange, " PUT "
	case "DELETE":
		return logger.ColorBrightWhite, logger.BgRed, " DEL "
	case "PATCH":
		return logger.ColorBlack, logger.BgMagenta, " PATCH "
	case "HEAD":
		return logger.ColorBlack, logger.BgWhite, " HEAD "
	case "OPTIONS":
		return logger.ColorBrightWhite, logger.BgPurple, " OPT "
	default:
		return logger.ColorBlack, logger.BgGray, " ??? "
	}
}

// Logger returns a Gin middleware that logs standardized information about every HTTP request.
// It captures the request method, final path (including query string), status code, and latency.
// Logging is performed asynchronously to ensure minimal impact on the request processing time.
func Logger(l logger.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Request ID Management
		// Ensure every request has a unique identifier for distributed tracing.
		requestID := c.GetHeader(consts.HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header(consts.HeaderXRequestID, requestID)
		c.Set(consts.CtxKeyRequestID, requestID)

		// Propagate RequestID to the standard context for deeper layers (Service/Repo)
		ctx := c.Request.Context()
		ctx = contextx.WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)

		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request_id", requestID))
		}

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Execute the handler chain.
		c.Next()

		// Construct the full path with query parameters for detailed tracing.
		if raw != "" {
			path = path + "?" + raw
		}

		// Create a shallow copy of the Gin context for use in the background goroutine.
		// THIS IS CRITICAL: Gin recycles context objects. Accessing the original context in a
		// goroutine after c.Next() returns can lead to race conditions and data corruption.
		cp := c.Copy()

		// Retrieve the correlation ID for logging.
		requestID = c.GetString(consts.CtxKeyRequestID)

		status := cp.Writer.Status()
		method := cp.Request.Method
		latency := time.Since(start)

		// Get background and foreground colors
		statusFg, statusBg := getStatusColor(status)
		methodFg, methodBg, methodLabel := getMethodStyle(method)

		// Format colored message - Matrix Style Badges
		coloredStatus := fmt.Sprintf("%s%s%s %d %s", statusBg, statusFg, logger.Bold, status, logger.ColorReset)
		coloredMethod := fmt.Sprintf("%s%s%s%s%s", methodBg, methodFg, logger.Bold, methodLabel, logger.ColorReset)

		// Path pops in White, Metadata/Query dims in Green
		var coloredPath string
		if raw != "" {
			coloredPath = fmt.Sprintf("%s%s%s%s %s%s#%s%s",
				logger.ColorBrightWhite, logger.Bold, c.Request.URL.Path, logger.ColorReset,
				logger.ColorGreen, logger.Dim+logger.Italic, raw, logger.ColorReset)
		} else {
			coloredPath = fmt.Sprintf("%s%s%s%s", logger.ColorBrightWhite, logger.Bold, path, logger.ColorReset)
		}

		coloredLatency := fmt.Sprintf("%s%s%s%s", logger.ColorGreen, logger.Dim, latency.String(), logger.ColorReset)

		// Dispatch logging to a separate goroutine.
		go func() {
			l.Infow(fmt.Sprintf("%s %s %s %s", coloredMethod, coloredPath, coloredStatus, coloredLatency),
				zap.String(consts.CtxKeyRequestID, requestID),
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
				zap.String("client_ip", httpx.GetIPAddress(cp)),
				zap.String("error", cp.Errors.String()),
			)
		}()
	}
}
