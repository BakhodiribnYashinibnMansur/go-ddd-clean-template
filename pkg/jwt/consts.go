package jwt

const (
	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Signing methods
	SigningMethodRS256 = "RS256"

	// PEM newline replacement
	PEMNewlineEscaped = "\\n"
	PEMNewline        = "\n"

	// Error messages
	ErrMsgParsePrivateKey = "jwt - ParseRSAPrivateKey - ParseRSAPrivateKeyFromPEM: %w"
	ErrMsgParsePublicKey  = "jwt - ParseRSAPublicKey - ParseRSAPublicKeyFromPEM: %w"

	// Token header keys
	HeaderAlg = "alg"
)
