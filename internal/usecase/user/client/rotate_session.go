package client

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/jwt"
	"gct/pkg/validator"
)

// RotateSession performs refresh token rotation by generating new tokens for an existing valid session.
func (uc *UseCase) RotateSession(ctx context.Context, in *domain.RefreshIn) (*domain.SignInOut, error) {
	uc.logger.Infoc(ctx, "session rotate started", "session_id", in.SessionID)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	sessionID := in.SessionID

	// 1. Get existing session
	session, err := uc.repo.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sessionID})
	if err != nil {
		uc.logger.Errorc(ctx, "session rotate failed: get session", "error", err)
		return nil, apperrors.MapRepoToServiceError(err)
	}

	// 3. Double check state
	if session.Revoked || session.IsExpired() {
		uc.logger.Warnc(ctx, "session rotate failed: session revoked or expired", "session_id", sessionID)
		return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "Session revoked or expired")
	}

	// 3. Generate NEW Refresh Token
	newRefToken, err := jwt.GenerateRefreshToken(
		session.UserID.String(),
		session.ID.String(),
		session.DeviceID.String(),
		uc.cfg.JWT.RefreshTTL,
	)
	if err != nil {
		uc.logger.Errorc(ctx, "session rotate failed: generate new refresh token", "error", err)
		return nil, apperrors.MapRepoToServiceError(err)
	}

	// 4. Update session in DB with new refresh token hash and extended expiry
	session.RefreshTokenHash = newRefToken.Hashed
	session.ExpiresAt = time.Now().Add(uc.cfg.JWT.RefreshTTL)
	session.UpdateActivity()

	err = uc.repo.Postgres.User.SessionRepo.Update(ctx, session)
	if err != nil {
		uc.logger.Errorc(ctx, "session rotate failed: update session", "error", err)
		return nil, apperrors.MapRepoToServiceError(err)
	}

	// 5. Generate NEW Access Token
	newAccessToken, err := jwt.GenerateToken(jwt.TokenParams{
		Issuer:     uc.cfg.JWT.Issuer,
		Subject:    session.UserID.String(),
		SessionID:  session.ID.String(),
		Type:       consts.TokenTypeAccess,
		TTL:        uc.cfg.JWT.AccessTTL,
		PrivateKey: uc.privateKey,
	})
	if err != nil {
		uc.logger.Errorc(ctx, "session rotate failed: generate access token", "error", err)
		return nil, apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infoc(ctx, "session rotate success", "user_id", session.UserID, "session_id", sessionID)
	return &domain.SignInOut{
		UserID:       session.UserID,
		SessionID:    session.ID,
		AccessToken:  newAccessToken,
		RefreshToken: newRefToken.String(),
	}, nil
}
