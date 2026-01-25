package client

import (
	"context"
	"strings"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/jwt"
	"gct/pkg/useragent"
	"gct/pkg/validator"

	"github.com/google/uuid"
)

func (uc *UseCase) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	uc.logger.Infoc(ctx, "user sign in started", "login", in.Login)

	// Validate input
	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	var user *domain.User
	var err error

	if in.Login != "" {
		// Simple heuristic: if it contains '@', it's an email, otherwise phone
		// This can be improved with proper validation if needed
		if strings.Contains(in.Login, "@") {
			user, err = uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{Email: &in.Login})
		} else {
			user, err = uc.repo.Postgres.User.Client.GetByPhone(ctx, in.Login)
		}
	} else {
		return nil, domain.ErrInvalidPassword
	}

	if err != nil {
		uc.logger.Errorc(ctx, "user sign in failed: get user", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	if !user.ComparePassword(in.Password) {
		err := domain.ErrInvalidPassword
		uc.logger.Errorc(ctx, "user sign in failed: invalid password", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	if !user.IsApproved {
		err := domain.ErrUserNotApproved
		uc.logger.Warnc(ctx, "user sign in failed: not approved", "user_id", user.ID)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	// Device Info
	deviceID := in.Session.DeviceID
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
		uc.logger.Errorc(ctx, "user sign in failed: generate refresh token", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	// Parse User-Agent
	ua := useragent.ParseUserAgent(in.Session.UserAgent)

	// Map device type
	var deviceType domain.SessionDeviceType
	switch ua.DeviceType {
	case useragent.DeviceTypeMobile:
		deviceType = domain.DeviceTypeMobile
	case useragent.DeviceTypeTablet:
		deviceType = domain.DeviceTypeTablet
	case useragent.DeviceTypeBot:
		deviceType = domain.DeviceTypeBot
	default:
		deviceType = domain.DeviceTypeDesktop
	}

	// Create clean device name from browser and OS
	deviceName := ua.Browser
	if ua.OS != "" {
		deviceName = ua.Browser + " on " + ua.OS
	}

	if in.Session.DeviceName != "" {
		deviceName = in.Session.DeviceName
	}
	if in.Session.DeviceType != "" {
		deviceType = domain.SessionDeviceType(in.Session.DeviceType)
	}
	if in.Session.OS != "" {
		ua.OS = in.Session.OS
	}
	if in.Session.OSVersion != "" {
		ua.OSVersion = in.Session.OSVersion
	}
	if in.Session.Browser != "" {
		ua.Browser = in.Session.Browser
	}
	if in.Session.BrowserVersion != "" {
		ua.BrowserVersion = in.Session.BrowserVersion
	}

	// Create session context with essential data
	sessionCtx := &domain.SessionContext{
		RoleID:   user.RoleID,
		Language: "uz", // default language, can be overridden by user preference
	}

	// Create Session
	session := &domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		DeviceID:         deviceID,
		DeviceName:       &deviceName,
		DeviceType:       &deviceType,
		UserAgent:        &in.Session.UserAgent,
		OS:               &ua.OS,
		OSVersion:        &ua.OSVersion,
		Browser:          &ua.Browser,
		BrowserVersion:   &ua.BrowserVersion,
		IPAddress:        &in.Session.IP,
		RefreshTokenHash: refToken.Hashed,
		ExpiresAt:        time.Now().Add(uc.cfg.JWT.RefreshTTL),
		CreatedAt:        time.Now(),
		LastActivity:     time.Now(),
	}

	// Set session context
	if err := session.SetContext(sessionCtx); err != nil {
		uc.logger.Errorc(ctx, "user sign in failed: set session context", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	err = uc.repo.Postgres.User.SessionRepo.Create(ctx, session)
	if err != nil {
		uc.logger.Errorc(ctx, "user sign in failed: create session", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
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
		uc.logger.Errorc(ctx, "user sign in failed: generate access token", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

	uc.logger.Infoc(ctx, "user sign in success", "user_id", user.ID, "session_id", sessionID)
	return &domain.SignInOut{
		UserID:       user.ID,
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}, nil
}
