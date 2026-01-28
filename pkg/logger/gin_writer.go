package logger

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ColorfulWriter wraps an io.Writer and adds colors to Gin debug logs
type ColorfulWriter struct {
	writer io.Writer
}

// NewColorfulWriter creates a new colorful writer
func NewColorfulWriter(w io.Writer) *ColorfulWriter {
	return &ColorfulWriter{writer: w}
}

// Write implements io.Writer interface with colorization
func (cw *ColorfulWriter) Write(p []byte) (n int, err error) {
	output := string(p)

	// Colorize [GIN-debug] prefix
	output = strings.ReplaceAll(output, "[GIN-debug]", ColorGreen+Dim+"[~] GIN_DEBUG"+ColorReset)
	output = strings.ReplaceAll(output, "[GIN]", ColorBrightGreen+"[+] GIN_CORE"+ColorReset)
	output = strings.ReplaceAll(output, "[WARNING]", ColorBrightYellow+Bold+"[!] WARNING"+ColorReset)
	output = strings.ReplaceAll(output, "[ERROR]", BgRed+ColorBrightWhite+Bold+" [X] ERROR "+ColorReset)

	// Colorize HTTP methods
	output = colorizeHTTPMethods(output)

	// Colorize routes (paths starting with /)
	output = colorizePaths(output)

	// Colorize router handlers (text after -->)
	output = colorizeRouters(output)

	// Colorize handler info (text in parentheses)
	output = colorizeHandlers(output)

	return cw.writer.Write([]byte(output))
}

// colorizeHTTPMethods adds colors with backgrounds to HTTP method names
func colorizeHTTPMethods(s string) string {
	methods := map[string]struct {
		bg    string
		fg    string
		label string
	}{
		"GET":     {BgCyan, ColorBlack, " GET "},
		"POST":    {BgGreen, ColorBlack, " POST "},
		"PUT":     {BgYellow, ColorBlack, " PUT "},
		"DELETE":  {BgRed, ColorWhite, " DEL "},
		"PATCH":   {BgMagenta, ColorWhite, " PATCH "},
		"HEAD":    {BgGray, ColorBlack, " HEAD "},
		"OPTIONS": {BgPurple, ColorWhite, " OPT "},
		"CONNECT": {BgPink, ColorBlack, " CONN "},
		"TRACE":   {BgOrange, ColorBlack, " TRACE "},
	}

	for method, style := range methods {
		// Match method followed by whitespace
		re := regexp.MustCompile(`\b` + method + `\s+`)
		s = re.ReplaceAllStringFunc(s, func(match string) string {
			return style.bg + style.fg + Bold + style.label + ColorReset + " "
		})
	}

	return s
}

// colorizePaths adds colors to URL paths
func colorizePaths(s string) string {
	// Match paths like /api/v1/users, /admin/dashboard, etc.
	re := regexp.MustCompile(`(/[a-zA-Z0-9/_:\-\*]+)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Clean White for paths in Matrix theme - pops against green
		return ColorBrightWhite + Bold + match + ColorReset
	})
}

// colorizeRouters adds colors to router handler names (text after -->)
func colorizeRouters(s string) string {
	// Match text after --> but before (
	re := regexp.MustCompile(`-->\s+([^\s(]+)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the handler name
		parts := strings.Split(match, "-->")
		if len(parts) == 2 {
			handler := strings.TrimSpace(parts[1])
			// Matrix Green Dim for handler details - reduces visual noise
			return " --> " + ColorGreen + Dim + Italic + handler + ColorReset
		}
		return match
	})
}

// colorizeHandlers adds colors to handler information
func colorizeHandlers(s string) string {
	// Match text in parentheses like (4 handlers)
	re := regexp.MustCompile(`\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Dim White/Gray for handler metadata
		return ColorGray + Dim + match + ColorReset
	})
}

// ColorizeGinOutput formats Gin startup messages with colors
func ColorizeGinOutput(message string) string {
	var result strings.Builder

	// Add box drawing characters and colors
	result.WriteString(ColorGray + "┌" + strings.Repeat("─", 60) + "┐" + ColorReset + "\n")

	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if line != "" {
			result.WriteString(ColorGray + "│ " + ColorReset)
			result.WriteString(ColorWhite + line + ColorReset)
			result.WriteString("\n")
		}
	}

	result.WriteString(ColorGray + "└" + strings.Repeat("─", 60) + "┘" + ColorReset + "\n")

	return result.String()
}

// PrintGinBanner prints a colorful banner for Gin startup
func PrintGinBanner(port int, mode string) {
	banner := fmt.Sprintf(`
%s┌───────────────────────────────────────────────────────────┐%s
%s│                                                           │%s
%s│     %s[ ACCESS_GRANTED ]%s                                   │%s
%s│     %sSYSTEM_READY: GIN_HTTP_SERVER_ON_LINE%s                │%s
%s│                                                           │%s
%s│     [ ADDRESS ]: %-6d                                   │%s
%s│     [ STATUS  ]: %s%-46s%s│%s
%s│                                                           │%s
%s└───────────────────────────────────────────────────────────┘%s
`,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorBrightGreen+Bold, ColorReset+ColorGreen, ColorReset,
		ColorGreen, ColorBrightGreen, ColorReset+ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, port, ColorReset,
		ColorGreen, ColorBrightCyan+Bold, mode, ColorReset+ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
	)

	fmt.Println(banner)
}
