package consts

// Response envelope status values. The presentation layer wraps all API responses in a
// { "status": "...", "message": "...", "data": ... } envelope using these constants.
const (
	ResponseStatusSuccess = "success"
	ResponseStatusError   = "error"
)

const (
	ResponseMessageSuccess      = "Success"
	ResponseMessageUnknownError = "Unknown error occurred"
)

const (
	TypeHandlerError = "handler_error"
)

// Signature verification and integration auth failure messages returned in error responses.
// These are human-readable; the corresponding error codes live in errors.go.
const (
	// Signature verification messages
	MsgSignTimeEmpty      = "time is empty"
	MsgSignInvalidTime    = "invalid time"
	MsgSignTimeExpired    = "time is expired"
	MsgSignRequestIDEmpty = "request id is empty"
	MsgSignEmpty          = "sign is empty"
	MsgSignInvalid        = "invalid sign"
	MsgSignPlatformEmpty  = "platform is empty"
	MsgSignInvalidPlat    = "invalid platform"

	// Integration Auth messages
	MsgIntegrationAuthReq = "integration authentication required"
	MsgInvalidAPIKey      = "invalid or expired api key"
	MsgMissingAPIKey      = "missing api key"
)
