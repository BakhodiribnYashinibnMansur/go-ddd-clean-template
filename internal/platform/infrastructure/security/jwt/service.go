package jwt

import (
	"crypto/rsa"
	"time"
)

// TokenService handles JWT token operations
type TokenService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   string
}

// NewTokenService creates a new token service
func NewTokenService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer, audience string) *TokenService {
	return &TokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		audience:   audience,
	}
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateTokenPair generates both access and refresh tokens
func (s *TokenService) GenerateTokenPair(userID, sessionID, clientID string) (*TokenPair, *RefreshToken, error) {
	// Generate access token (15 minutes)
	accessToken, err := GenerateAccessToken(
		userID,
		sessionID,
		s.issuer,
		s.audience,
		s.privateKey,
		15*time.Minute,
	)
	if err != nil {
		return nil, nil, err
	}

	// Generate refresh token (7 days)
	refreshToken, err := GenerateRefreshToken(
		userID,
		sessionID,
		clientID,
		7*24*time.Hour,
	)
	if err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String(),
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes in seconds
	}, refreshToken, nil
}

// ValidateAccessToken validates an access token
func (s *TokenService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	return ParseAccessToken(tokenString, s.publicKey, s.issuer, s.audience)
}

// ValidateRefreshToken validates a refresh token
func (s *TokenService) ValidateRefreshToken(tokenString, storedHash string) (*RefreshToken, error) {
	token, err := VerifyRefreshToken(tokenString, storedHash)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// RefreshTokens generates new tokens using a valid refresh token
func (s *TokenService) RefreshTokens(refreshTokenString, storedHash string) (*TokenPair, *RefreshToken, error) {
	// Verify the refresh token
	refreshToken, err := s.ValidateRefreshToken(refreshTokenString, storedHash)
	if err != nil {
		return nil, nil, err
	}

	// Check if the refresh token is expired
	if refreshToken.IsExpired() {
		return nil, nil, ErrRefreshTokenExpired
	}

	// Generate new tokens
	return s.GenerateTokenPair(
		refreshToken.UserID,
		refreshToken.SessionID,
		refreshToken.ClientID,
	)
}
