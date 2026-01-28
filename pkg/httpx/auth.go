package httpx

import (
	"strings"

	"gct/consts"
)

// ExtractBearerToken parses the "Bearer <token>" string from an authorization header value.
//
// Returns empty string if the authorization header is missing or malformed.
// Expected format: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
//
// Example:
//
//	authHeader := "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
//	token := httpx.ExtractBearerToken(authHeader)
//	// token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
func ExtractBearerToken(authHeader string) string {
	if authHeader == EmptyString {
		return EmptyString
	}

	parts := strings.Split(authHeader, SeparatorSpace)

	if len(parts) != ExpectedAuthParts || parts[0] != consts.BearerToken {
		return EmptyString
	}

	return parts[1]
}

// ExtractBasicToken parses the "Basic <token>" string from an authorization header value.
//
// Returns empty string if the authorization header is missing or malformed.
// Expected format: "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
//
// Example:
//
//	authHeader := "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
//	token := httpx.ExtractBasicToken(authHeader)
//	// token = "dXNlcm5hbWU6cGFzc3dvcmQ="
func ExtractBasicToken(authHeader string) string {
	if authHeader == EmptyString {
		return EmptyString
	}

	parts := strings.Split(authHeader, SeparatorSpace)

	if len(parts) != ExpectedAuthParts || parts[0] != consts.BasicToken {
		return EmptyString
	}

	return parts[1]
}

// ParseAuthorizationType determines the type of authorization from the header value.
//
// Returns "bearer", "basic", or empty string if unrecognized.
//
// Example:
//
//	authType := httpx.ParseAuthorizationType("Bearer token123")
//	// authType = "bearer"
func ParseAuthorizationType(authHeader string) string {
	if authHeader == EmptyString {
		return EmptyString
	}

	parts := strings.Split(authHeader, SeparatorSpace)
	if len(parts) < MinAuthParts {
		return EmptyString
	}

	authType := strings.ToLower(parts[0])
	switch authType {
	case strings.ToLower(consts.BearerToken):
		return AuthTypeBearer
	case strings.ToLower(consts.BasicToken):
		return AuthTypeBasic
	default:
		return EmptyString
	}
}
