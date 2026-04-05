package telegram

import "time"

const (
	// DefaultTimeout is the default timeout for the HTTP client
	DefaultTimeout = 10 * time.Second

	// APIURLFormat is the format string for the Telegram API URL
	APIURLFormat = "https://api.telegram.org/bot%s/sendMessage"

	// Message prefixes
	PrefixError = "🚨 Error: "
	PrefixInfo  = "ℹ️ Info: "

	// Content Types
	ContentTypeJSON = "application/json"
)
