package httpx

const (
	// Authorization types
	AuthTypeBearer = "Bearer"
	AuthTypeBasic  = "Basic"

	// Common separators
	SeparatorSpace = " "

	// Empty string
	EmptyString = ""

	// Number constants
	ExpectedAuthParts = 2
	MinAuthParts      = 1

	// HTTP Headers
	HeaderContentDescription = "Content-Description"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentType        = "Content-Type"

	// Content Types
	ContentTypeOctetStream = "application/octet-stream"

	// File transfer
	FileTransferDescription = "File Transfer"
	AttachmentPrefix        = "attachment; filename="
	CurrentDir              = "./"
)
