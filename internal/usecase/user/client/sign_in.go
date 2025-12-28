package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/google/uuid"
)

func (uc *UseCase) SignIn(ctx context.Context, in SignInInput) (SignInOutput, error) {
	user, err := uc.repo.User.Client.GetByPhone(ctx, in.Phone)
	if err != nil {
		return SignInOutput{}, fmt.Errorf("user not found: %w", err)
	}

	if !user.ComparePassword(in.Password) {
		return SignInOutput{}, fmt.Errorf("invalid password")
	}

	// Device Info
	deviceID, err := uuid.Parse(in.DeviceID)
	if err != nil {
		deviceID = uuid.New()
	}

	sessionID := uuid.New()

	// Generate Refresh Token
	refToken, err := jwt.GenerateRefreshToken(
		strconv.FormatInt(user.ID, 10),
		sessionID.String(),
		deviceID.String(),
		uc.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		return SignInOutput{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create Session
	session := &domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		DeviceID:         deviceID,
		DeviceName:       &in.UserAgent,
		IPAddress:        &in.IP,
		RefreshTokenHash: refToken.Hashed,
	}

	err = uc.repo.User.SessionRepo.Create(ctx, session)
	if err != nil {
		return SignInOutput{}, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate Access Token
	accessToken, err := jwt.GenerateToken(jwt.TokenParams{
		Issuer:     uc.cfg.JWT.Issuer,
		Subject:    strconv.FormatInt(user.ID, 10),
		SessionID:  sessionID.String(),
		Type:       "access",
		TTL:        uc.cfg.JWT.AccessTTL,
		PrivateKey: uc.privateKey,
	})
	if err != nil {
		return SignInOutput{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	return SignInOutput{
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}, nil
}
