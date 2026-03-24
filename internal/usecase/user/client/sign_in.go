package client

import (
	"context"
	"strings"
	"time"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/shared/infrastructure/ptrutil"
	"gct/internal/shared/infrastructure/useragent"
	"gct/internal/shared/infrastructure/validator"

	"github.com/google/uuid"
)

func (uc *UseCase) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	login := ptrutil.StrVal(in.Login)
	password := ptrutil.StrVal(in.Password)

	uc.logger.Infoc(ctx, "user sign in started", "login", login)

	if err := validator.ValidateStruct(in); err != nil {
		return nil, err
	}

	user, err := uc.resolveUser(ctx, login)
	if err != nil {
		uc.logger.Errorc(ctx, "user sign in failed: get user", "error", err)
		return nil, domain.ErrInvalidPassword
	}

	if !user.ComparePassword(password) {
		uc.logger.Errorc(ctx, "user sign in failed: invalid password")
		return nil, apperrors.MapRepoToServiceError(domain.ErrInvalidPassword).WithInput(in)
	}

	if !user.IsApproved {
		uc.logger.Warnc(ctx, "user sign in failed: not approved", "user_id", user.ID)
		return nil, apperrors.MapRepoToServiceError(domain.ErrUserNotApproved).WithInput(in)
	}

	deviceID := in.Session.DeviceID
	if deviceID == uuid.Nil {
		deviceID = uuid.New()
	}
	sessionID := uuid.New()

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

	deviceName, deviceType, ua := uc.mapDeviceInfo(in.Session)

	session := buildSession(sessionID, user.ID, deviceID, deviceName, deviceType, ua, in.Session, refToken.Hashed, uc.cfg.JWT.RefreshTTL)

	err = uc.repo.Postgres.User.SessionRepo.Create(ctx, session)
	if err != nil {
		uc.logger.Errorc(ctx, "user sign in failed: create session", "error", err)
		return nil, apperrors.MapRepoToServiceError(err).WithInput(in)
	}

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

// resolveUser looks up a user by email or phone based on a simple heuristic.
func (uc *UseCase) resolveUser(ctx context.Context, login string) (*domain.User, error) {
	if login == "" {
		return nil, domain.ErrInvalidPassword
	}
	if strings.Contains(login, "@") {
		return uc.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{Email: &login})
	}
	return uc.repo.Postgres.User.Client.GetByPhone(ctx, login)
}

// mapDeviceInfo parses user-agent and session overrides into device metadata.
func (uc *UseCase) mapDeviceInfo(session *domain.SessionIn) (deviceName string, deviceType domain.SessionDeviceType, ua *useragent.UserAgent) {
	ua = useragent.ParseUserAgent(session.UserAgent)

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

	deviceName = ua.Browser
	if ua.OS != "" {
		deviceName = ua.Browser + " on " + ua.OS
	}

	// Apply client-provided overrides
	if session.DeviceName != "" {
		deviceName = session.DeviceName
	}
	if session.DeviceType != "" {
		deviceType = domain.SessionDeviceType(strings.ToUpper(session.DeviceType))
	}
	if session.OS != "" {
		ua.OS = session.OS
	}
	if session.OSVersion != "" {
		ua.OSVersion = session.OSVersion
	}
	if session.Browser != "" {
		ua.Browser = session.Browser
	}
	if session.BrowserVersion != "" {
		ua.BrowserVersion = session.BrowserVersion
	}

	return deviceName, deviceType, ua
}

// buildSession constructs a domain.Session from the provided parameters.
func buildSession(
	sessionID, userID, deviceID uuid.UUID,
	deviceName string,
	deviceType domain.SessionDeviceType,
	ua *useragent.UserAgent,
	sessionIn *domain.SessionIn,
	refreshTokenHash string,
	refreshTTL time.Duration,
) *domain.Session {
	now := time.Now()

	// Validate device type against allowed ENUM values; default to DESKTOP
	if !deviceType.IsValid() {
		deviceType = domain.DeviceTypeDesktop
	}

	// PostgreSQL inet type cannot accept empty string; use nil instead
	var ipAddr *string
	if sessionIn.IP != "" {
		ipAddr = &sessionIn.IP
	}

	return &domain.Session{
		ID:               sessionID,
		UserID:           userID,
		DeviceID:         deviceID,
		DeviceName:       &deviceName,
		DeviceType:       &deviceType,
		UserAgent:        &sessionIn.UserAgent,
		OS:               &ua.OS,
		OSVersion:        &ua.OSVersion,
		Browser:          &ua.Browser,
		BrowserVersion:   &ua.BrowserVersion,
		IPAddress:        ipAddr,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        now.Add(refreshTTL),
		CreatedAt:        now,
		LastActivity:     now,
	}
}
