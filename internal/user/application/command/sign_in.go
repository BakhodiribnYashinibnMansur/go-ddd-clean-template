package command

import (
	"context"
	"crypto/rsa"
	"strings"
	"time"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	jwtpkg "gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// SignInCommand represents an authentication attempt using login credentials and device metadata.
// Login accepts either a phone number or email — the handler auto-detects the format via "@" presence.
// DeviceType is uppercased internally to match the PostgreSQL ENUM constraint (e.g., "WEB", "MOBILE").
type SignInCommand struct {
	Login      string
	Password   string
	DeviceType string
	IP         string
	UserAgent  string
}

// SignInResult holds the output of a successful sign-in.
type SignInResult struct {
	UserID       uuid.UUID
	SessionID    uuid.UUID
	AccessToken  string
	RefreshToken string
}

// JWTConfig holds the parameters needed for JWT token generation.
type JWTConfig struct {
	PrivateKey *rsa.PrivateKey
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// SignInHandler handles the SignInCommand.
type SignInHandler struct {
	repo      domain.UserRepository
	eventBus  application.EventBus
	logger    logger.Log
	signIn    domain.SignInService
	jwtConfig JWTConfig
}

// NewSignInHandler creates a new SignInHandler.
func NewSignInHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
	jwtCfg JWTConfig,
) *SignInHandler {
	return &SignInHandler{
		repo:      repo,
		eventBus:  eventBus,
		logger:    logger,
		signIn:    domain.SignInService{},
		jwtConfig: jwtCfg,
	}
}

// Handle executes the SignInCommand and returns SignInResult.
func (h *SignInHandler) Handle(ctx context.Context, cmd SignInCommand) (result *SignInResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignInHandler.Handle")
	defer func() { end(err) }()

	// Find user by phone or email based on login format.
	user, err := h.findUser(ctx, cmd.Login)
	if err != nil {
		return nil, err
	}

	deviceType := domain.SessionDeviceType(strings.ToUpper(cmd.DeviceType))

	session, err := h.signIn.SignIn(user, cmd.Password, deviceType, cmd.IP, cmd.UserAgent)
	if err != nil {
		return nil, err
	}

	// Generate refresh token and store its hash on the session before persisting.
	refToken, err := jwtpkg.GenerateRefreshToken(
		user.ID().String(),
		session.ID().String(),
		session.DeviceID(),
		h.jwtConfig.RefreshTTL,
	)
	if err != nil {
		h.logger.Errorf("failed to generate refresh token: %v", err)
		return nil, err
	}
	session.SetRefreshTokenHash(refToken.Hashed)

	// Persist user (with the updated session containing the refresh token hash).
	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save user after sign-in: %v", err)
		return nil, err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish sign-in events: %v", err)
	}

	// Generate access token (signed JWT).
	accessToken, err := jwtpkg.GenerateAccessToken(
		user.ID().String(),
		session.ID().String(),
		h.jwtConfig.Issuer,
		"", // audience
		h.jwtConfig.PrivateKey,
		h.jwtConfig.AccessTTL,
	)
	if err != nil {
		h.logger.Errorf("failed to generate access token: %v", err)
		return nil, err
	}

	result = &SignInResult{
		UserID:       user.ID(),
		SessionID:    session.ID(),
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}

	return result, nil
}

// findUser looks up a user by phone or email depending on the login format.
func (h *SignInHandler) findUser(ctx context.Context, login string) (*domain.User, error) {
	if strings.Contains(login, "@") {
		email, err := domain.NewEmail(login)
		if err != nil {
			return nil, err
		}
		return h.repo.FindByEmail(ctx, email)
	}

	phone, err := domain.NewPhone(login)
	if err != nil {
		return nil, err
	}
	return h.repo.FindByPhone(ctx, phone)
}
