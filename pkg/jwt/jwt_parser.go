package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/evrone/go-clean-template/consts"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// TokenMetadata describes the metadata extracted from a JWT.
type TokenMetadata struct {
	Issuer          string
	Subject         string
	SessionID       string
	CompanyID       string
	Audience        string
	Scope           []string
	AuthorizedParty string
	Type            string
	ExpiresAt       int64
	IssuedAt        int64
	JWTID           string
}

// ParseToken parses and validates a JWT token string.
func ParseToken(tokenString string, publicKey *rsa.PublicKey) (*TokenMetadata, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Helper to extract string claims safely
	getString := func(key string) string {
		if val, exists := claims[key]; exists {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return ""
	}

	// Helper to extract int64 claims safely
	getInt64 := func(key string) int64 {
		if val, exists := claims[key]; exists {
			if f, ok := val.(float64); ok {
				return int64(f)
			}
		}
		return 0
	}

	// Helper to extract string slice claims safely
	getStringSlice := func(key string) []string {
		if val, exists := claims[key]; exists {
			if i, ok := val.([]interface{}); ok {
				s := make([]string, len(i))
				for idx, v := range i {
					if str, ok := v.(string); ok {
						s[idx] = str
					}
				}
				return s
			}
		}
		return nil
	}

	metadata := &TokenMetadata{
		Issuer:          getString(consts.ClaimIssuer),
		Subject:         getString(consts.ClaimSubject),
		SessionID:       getString(consts.ClaimSessionID),
		CompanyID:       getString(consts.ClaimCompanyID),
		Audience:        getString(consts.ClaimAudience),
		Scope:           getStringSlice(consts.ClaimScope),
		AuthorizedParty: getString(consts.ClaimAuthorizedParty),
		Type:            getString(consts.ClaimType),
		ExpiresAt:       getInt64(consts.ClaimExpiresAt),
		IssuedAt:        getInt64(consts.ClaimIssuedAt),
		JWTID:           getString(consts.ClaimJWTID),
	}

	if metadata.ExpiresAt < time.Now().Unix() {
		return nil, ErrTokenExpired
	}

	return metadata, nil
}
