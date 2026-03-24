package command

import (
	"context"
	"strings"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// SignInCommand holds the input for user sign-in.
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

// SignInHandler handles the SignInCommand.
type SignInHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
	signIn   domain.SignInService
}

// NewSignInHandler creates a new SignInHandler.
func NewSignInHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *SignInHandler {
	return &SignInHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
		signIn:   domain.SignInService{},
	}
}

// Handle executes the SignInCommand and returns SignInResult.
func (h *SignInHandler) Handle(ctx context.Context, cmd SignInCommand) (*SignInResult, error) {
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

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save user after sign-in: %v", err)
		return nil, err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish sign-in events: %v", err)
	}

	// TODO: Generate JWT access and refresh tokens here.
	// Actual JWT generation will be wired in Plan 6 when integrating with the existing JWT package.
	result := &SignInResult{
		UserID:       user.ID(),
		SessionID:    session.ID(),
		AccessToken:  "", // TODO: generate JWT access token
		RefreshToken: "", // TODO: generate JWT refresh token
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
