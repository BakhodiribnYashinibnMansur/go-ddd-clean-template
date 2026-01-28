package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	ErrAccessTokenInvalid = errors.New("invalid access token")
	ErrAccessTokenExpired = errors.New("access token has expired")
)

// AccessTokenClaims represents claims for access tokens
type AccessTokenClaims struct {
	UserID    string `json:"sub"`
	SessionID string `json:"sid"`
	Type      string `json:"typ"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func GenerateAccessToken(userID, sessionID, issuer, audience string, privateKey *rsa.PrivateKey, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		Type:      TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// ParseAccessToken parses and validates an access token
func ParseAccessToken(tokenString string, publicKey *rsa.PublicKey, issuer, audience string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrAccessTokenInvalid, token.Header[HeaderAlg])
		}
		return publicKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrAccessTokenExpired
			}
		}
		return nil, fmt.Errorf("%w: %w", ErrAccessTokenInvalid, err)
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		// Validate issuer and audience
		if !claims.VerifyIssuer(issuer, true) {
			return nil, fmt.Errorf("%w: invalid issuer", ErrAccessTokenInvalid)
		}
		if !claims.VerifyAudience(audience, audience != "") {
			return nil, fmt.Errorf("%w: invalid audience", ErrAccessTokenInvalid)
		}
		return claims, nil
	}

	return nil, ErrAccessTokenInvalid
}
