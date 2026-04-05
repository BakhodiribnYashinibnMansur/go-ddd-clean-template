package jwt

import "time"

const (
	// Token type claim values (the "typ" custom claim on our JWT body).
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// SigningMethod is the only JWT algorithm accepted by this package.
	// RS256 is enforced strictly by the parser to prevent alg-downgrade attacks.
	SigningMethod = "RS256"

	// JWT header keys.
	HeaderKid = "kid"

	// Refresh-token wire format: <RefreshTokenPrefix><RefreshTokenVersion>.<sid>.<id>.<secret>
	RefreshTokenPrefix  = "rft_"
	RefreshTokenVersion = "v1"

	// Domain-separation prefix for the refresh-token HMAC.
	// Changing this invalidates all outstanding refresh tokens — use for cryptographic hygiene only.
	refreshHashDomain = "gct:rt:v1"

	// refreshSecretBytes is the raw-byte length of the random refresh-token secret (256 bits).
	refreshSecretBytes = 32
	// refreshTokenIDBytes is the raw-byte length of the random refresh-token ID (192 bits).
	refreshTokenIDBytes = 24

	// DefaultLeeway is applied when no clock-skew tolerance is supplied.
	DefaultLeeway = 30 * time.Second
)
