package jwt

import (
	"crypto/rsa"
	"time"

	"gct/internal/shared/domain/consts"
	jwt "github.com/golang-jwt/jwt/v4"
)

// Tokens represents a pair of access and refresh tokens.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenParams defines the parameters for generating a JWT.
type TokenParams struct {
	Issuer          string
	Subject         string
	SessionID       string
	CompanyID       string
	Audience        string
	Scope           []string
	AuthorizedParty string
	Type            string // "access" or "refresh"
	TTL             time.Duration
	PrivateKey      *rsa.PrivateKey
}

// GenerateToken generates a new JWT token with the specified params.
// It only includes optional claims (cid, scp, azp) if they are provided.
func GenerateToken(p TokenParams) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		consts.ClaimIssuer:    p.Issuer,
		consts.ClaimSubject:   p.Subject,
		consts.ClaimSessionID: p.SessionID,
		consts.ClaimAudience:  p.Audience,
		consts.ClaimType:      p.Type,
		consts.ClaimExpiresAt: now.Add(p.TTL).Unix(),
		consts.ClaimIssuedAt:  now.Unix(),
	}

	if p.CompanyID != "" {
		claims[consts.ClaimCompanyID] = p.CompanyID
	}
	if p.Scope != nil {
		claims[consts.ClaimScope] = p.Scope
	}
	if p.AuthorizedParty != "" {
		claims[consts.ClaimAuthorizedParty] = p.AuthorizedParty
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(p.PrivateKey)
}
