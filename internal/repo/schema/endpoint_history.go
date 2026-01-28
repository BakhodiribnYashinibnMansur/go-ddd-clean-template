package schema

// Table name
const TableEndpointHistory = "endpoint_history"

// EndpointHistory table columns
const (
	EndpointHistoryID           = "id"
	EndpointHistoryUserID       = "user_id"
	EndpointHistorySessionID    = "session_id"
	EndpointHistoryMethod       = "method"
	EndpointHistoryPath         = "path"
	EndpointHistoryStatusCode   = "status_code"
	EndpointHistoryDurationMs   = "duration_ms"
	EndpointHistoryPlatform     = "platform"
	EndpointHistoryIPAddress    = "ip_address"
	EndpointHistoryUserAgent    = "user_agent"
	EndpointHistoryPermission   = "permission"
	EndpointHistoryDecision     = "decision"
	EndpointHistoryRequestID    = "request_id"
	EndpointHistoryRateLimited  = "rate_limited"
	EndpointHistoryResponseSize = "response_size"
	EndpointHistoryErrorMessage = "error_message"
	EndpointHistoryCreatedAt    = "created_at"
)
