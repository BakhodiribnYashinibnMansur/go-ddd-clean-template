package client

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/jwt"
	"gct/pkg/validator"

	"github.com/google/uuid"
)

func (uc *UseCase) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	uc.logger.WithContext(ctx).Infow("user sign in started", "phone", in.Phone)

	// Validate input
	if err := validator.ValidateStruct(ctx, in); err != nil {
		return nil, err
	}

	user, err := uc.repo.Postgres.User.Client.GetByPhone(ctx, in.Phone)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("user sign in failed: get by phone", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	if !user.ComparePassword(in.Password) {
		err := domain.ErrInvalidPassword
		uc.logger.WithContext(ctx).Errorw("user sign in failed: invalid password", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	// Device Info
	deviceID := in.DeviceID
	if deviceID == uuid.Nil {
		deviceID = uuid.New()
	}

	sessionID := uuid.New()

	// Generate Refresh Token
	refToken, err := jwt.GenerateRefreshToken(
		user.ID.String(),
		sessionID.String(),
		deviceID.String(),
		uc.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("user sign in failed: generate refresh token", "error", err)
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
		ExpiresAt:        time.Now().Add(uc.cfg.JWT.RefreshTTL),
	}

	err = uc.repo.Postgres.User.SessionRepo.Create(ctx, session)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("user sign in failed: create session", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	// Generate Access Token
	accessToken, err := jwt.GenerateToken(jwt.TokenParams{
		Issuer:     uc.cfg.JWT.Issuer,
		Subject:    user.ID.String(),
		SessionID:  sessionID.String(),
		Type:       "access",
		TTL:        uc.cfg.JWT.AccessTTL,
		PrivateKey: uc.privateKey,
	})
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("user sign in failed: generate access token", "error", err)
		return nil, apperrors.MapRepoToServiceError(ctx, err).WithInput(in)
	}

	uc.logger.WithContext(ctx).Infow("user sign in success", "user_id", user.ID, "session_id", sessionID)
	return &domain.SignInOut{
		UserID:       user.ID,
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}, nil
}
