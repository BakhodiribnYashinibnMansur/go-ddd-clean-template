package logger

// ANSI color codes for terminal output
const (
	// Foreground colors
	ColorRed     = "\x1b[31m"
	ColorGreen   = "\x1b[32m"
	ColorYellow  = "\x1b[33m"
	ColorBlue    = "\x1b[34m"
	ColorMagenta = "\x1b[35m"
	ColorCyan    = "\x1b[36m"
	ColorWhite   = "\x1b[37m"
	ColorBlack   = "\x1b[30m"
	ColorGray    = "\x1b[90m" // Bright black

	// Bright foreground colors (more vibrant)
	ColorBrightRed     = "\x1b[91m"
	ColorBrightGreen   = "\x1b[92m"
	ColorBrightYellow  = "\x1b[93m"
	ColorBrightBlue    = "\x1b[94m"
	ColorBrightMagenta = "\x1b[95m"
	ColorBrightCyan    = "\x1b[96m"
	ColorBrightWhite   = "\x1b[97m"

	// Special colors (256 color palette)
	ColorOrange = "\x1b[38;5;208m" // Orange
	ColorPurple = "\x1b[38;5;141m" // Purple
	ColorPink   = "\x1b[38;5;213m" // Pink

	// Background colors
	BgBlack         = "\x1b[40m"
	BgRed           = "\x1b[41m"
	BgGreen         = "\x1b[42m"
	BgYellow        = "\x1b[43m"
	BgBlue          = "\x1b[44m"
	BgMagenta       = "\x1b[45m"
	BgCyan          = "\x1b[46m"
	BgWhite         = "\x1b[47m"
	BgGray          = "\x1b[100m"
	BgBrightRed     = "\x1b[101m"
	BgBrightGreen   = "\x1b[102m"
	BgBrightYellow  = "\x1b[103m"
	BgBrightBlue    = "\x1b[104m"
	BgBrightMagenta = "\x1b[105m"
	BgBrightCyan    = "\x1b[106m"
	BgBrightWhite   = "\x1b[107m"

	// Special Backgrounds (256 color palette)
	BgOrange = "\x1b[48;5;208m"
	BgPurple = "\x1b[48;5;141m"
	BgPink   = "\x1b[48;5;213m"

	// Text formatting
	Bold       = "\x1b[1m"
	Dim        = "\x1b[2m"
	Italic     = "\x1b[3m"
	Underline  = "\x1b[4m"
	ColorReset = "\x1b[0m"

	// Console separator - refined gray
	ConsoleSeparator = "\x1b[90m │ \x1b[0m"

	// Time format
	TimeFormat = "2006-01-02 15:04:05"
)

// Log level display formats
const (
	LevelDebugDisplay = "DEBUG"
	LevelInfoDisplay  = " INFO "
	LevelWarnDisplay  = " WARN "
	LevelErrorDisplay = " ERROR "
)
