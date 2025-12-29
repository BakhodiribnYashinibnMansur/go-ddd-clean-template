package client

import (
	"context"
	"strconv"

	"github.com/google/uuid"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/jwt"
)

func (uc *UseCase) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	user, err := uc.repo.Postgres.Client.GetByPhone(ctx, in.Phone)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	if !user.ComparePassword(in.Password) {
		return nil, apperrors.MapRepoToServiceError(ctx, domain.ErrInvalidPassword).WithInput(in)
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
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
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

	err = uc.repo.Postgres.SessionRepo.Create(ctx, session)
	if err != nil {
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
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
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	return &domain.SignInOut{
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}, nil
}
