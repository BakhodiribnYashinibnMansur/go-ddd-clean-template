package middleware

import (
	"fmt"
	"time"

	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/contextx"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/logger"

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
		requestID := ensureRequestID(c)

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
		status := cp.Writer.Status()
		method := cp.Request.Method
		latency := time.Since(start)

		coloredMsg := formatAccessLogMessage(c, path, raw, method, status, latency)

		// Dispatch logging to a separate goroutine.
		go func() {
			l.Infow(coloredMsg,
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

// ensureRequestID reads the X-Request-ID header, generates one if missing, and
// propagates it into the Gin context, response header, and request context for
// downstream layers (including the OTel span).
func ensureRequestID(c *gin.Context) string {
	requestID := c.GetHeader(consts.HeaderXRequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Header(consts.HeaderXRequestID, requestID)
	c.Set(consts.CtxKeyRequestID, requestID)

	ctx := contextx.WithRequestID(c.Request.Context(), requestID)
	c.Request = c.Request.WithContext(ctx)

	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("request_id", requestID))
	}
	return requestID
}

// formatAccessLogMessage builds the colored single-line access log message
// (method badge + path + status badge + latency).
func formatAccessLogMessage(c *gin.Context, path, raw, method string, status int, latency time.Duration) string {
	statusFg, statusBg := getStatusColor(status)
	methodFg, methodBg, methodLabel := getMethodStyle(method)

	coloredStatus := fmt.Sprintf("%s%s%s %d %s", statusBg, statusFg, logger.Bold, status, logger.ColorReset)
	coloredMethod := fmt.Sprintf("%s%s%s%s%s", methodBg, methodFg, logger.Bold, methodLabel, logger.ColorReset)

	var coloredPath string
	if raw != "" {
		coloredPath = fmt.Sprintf("%s%s%s%s %s%s#%s%s",
			logger.ColorBrightWhite, logger.Bold, c.Request.URL.Path, logger.ColorReset,
			logger.ColorGreen, logger.Dim+logger.Italic, raw, logger.ColorReset)
	} else {
		coloredPath = fmt.Sprintf("%s%s%s%s", logger.ColorBrightWhite, logger.Bold, path, logger.ColorReset)
	}

	coloredLatency := fmt.Sprintf("%s%s%s%s", logger.ColorGreen, logger.Dim, latency.String(), logger.ColorReset)

	return fmt.Sprintf("%s %s %s %s", coloredMethod, coloredPath, coloredStatus, coloredLatency)
}
