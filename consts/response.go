package consts

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
