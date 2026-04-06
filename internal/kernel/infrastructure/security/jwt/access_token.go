package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrAccessTokenInvalid is returned for any parse/validation failure
	// that is not specifically an expiry.
	ErrAccessTokenInvalid = errors.New("invalid access token")
	// ErrAccessTokenExpired is returned when the access token's "exp" claim
	// (with the configured leeway applied) is in the past.
	ErrAccessTokenExpired = errors.New("access token has expired")
)

// AccessTokenClaims is the body of an access token. It embeds the RFC 7519
// registered claims so the v5 parser can validate iss/aud/exp/iat/nbf, and
// adds custom claims: the session ID ("sid"), token type ("typ"), and an
// optional Token-Binding Hash ("tbh") for device-binding verification.
type AccessTokenClaims struct {
	SessionID string `json:"sid"`
	Type      string `json:"typ"`
	TBH       string `json:"tbh,omitempty"`
	jwtgo.RegisteredClaims
}

// GenerateAccessToken signs an RS256 JWT access token.
// All timestamps are emitted in UTC. If keyID is non-empty it is placed in
// the JWT header under "kid" to support future key rotation without a
// JWKS endpoint.
//
// An optional tbh (Token-Binding Hash) may be supplied as the last variadic
// argument. When non-empty it is embedded in the "tbh" claim so the
// middleware can verify the token is being used from the same device context.
func GenerateAccessToken(
	userID, sessionID, issuer, audience, keyID string,
	privateKey *rsa.PrivateKey,
	expiresIn time.Duration,
	opts ...string,
) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("jwt.GenerateAccessToken: private key is nil")
	}
	if audience == "" {
		return "", fmt.Errorf("jwt.GenerateAccessToken: audience is required")
	}
	if issuer == "" {
		return "", fmt.Errorf("jwt.GenerateAccessToken: issuer is required")
	}

	now := time.Now().UTC()
	claims := AccessTokenClaims{
		SessionID: sessionID,
		Type:      TokenTypeAccess,
		RegisteredClaims: jwtgo.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			Audience:  jwtgo.ClaimStrings{audience},
			ExpiresAt: jwtgo.NewNumericDate(now.Add(expiresIn)),
			NotBefore: jwtgo.NewNumericDate(now),
			IssuedAt:  jwtgo.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	if len(opts) > 0 && opts[0] != "" {
		claims.TBH = opts[0]
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, claims)
	if keyID != "" {
		token.Header[HeaderKid] = keyID
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("jwt.GenerateAccessToken: sign: %w", err)
	}
	return signed, nil
}

// ParseAccessToken verifies the signature and all registered claims of an
// access-token JWT using v5 parser options:
//   - strict RS256 (rejects alg-downgrade and "none")
//   - iss and aud must match exactly (non-empty)
//   - exp is required and validated with leeway
//   - iat is validated with leeway
//
// It additionally validates the custom "typ" claim equals "access".
func ParseAccessToken(
	tokenString string,
	publicKey *rsa.PublicKey,
	issuer, audience string,
	leeway time.Duration,
) (*AccessTokenClaims, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("jwt.ParseAccessToken: public key is nil")
	}
	if leeway <= 0 {
		leeway = DefaultLeeway
	}

	parser := jwtgo.NewParser(
		jwtgo.WithValidMethods([]string{SigningMethod}),
		jwtgo.WithIssuer(issuer),
		jwtgo.WithAudience(audience),
		jwtgo.WithExpirationRequired(),
		jwtgo.WithIssuedAt(),
		jwtgo.WithLeeway(leeway),
	)

	token, err := parser.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(t *jwtgo.Token) (any, error) {
		// The WithValidMethods option already vets t.Method; we need only
		// supply the verification key.
		return publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwtgo.ErrTokenExpired) {
			return nil, ErrAccessTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrAccessTokenInvalid, err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrAccessTokenInvalid
	}
	if claims.Type != TokenTypeAccess {
		return nil, fmt.Errorf("%w: wrong token type %q", ErrAccessTokenInvalid, claims.Type)
	}
	return claims, nil
}
