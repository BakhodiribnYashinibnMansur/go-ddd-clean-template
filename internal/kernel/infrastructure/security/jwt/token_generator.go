// This file previously held a second, duplicate token generator using
// jwt.MapClaims. It was dead code (zero production call sites) and was
// removed during the security refactor. GenerateAccessToken in
// access_token.go is the single canonical path.
package jwt
