package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/domain"
)

// SignUpCommand holds the input for user self-registration.
type SignUpCommand struct {
	Phone    string
	Password string
	Username *string
	Email    *string
}

// SignUpHandler handles the SignUpCommand.
type SignUpHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewSignUpHandler creates a new SignUpHandler.
func NewSignUpHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *SignUpHandler {
	return &SignUpHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the SignUpCommand.
// The user is created as active but NOT approved by default.
func (h *SignUpHandler) Handle(ctx context.Context, cmd SignUpCommand) error {
	phone, err := domain.NewPhone(cmd.Phone)
	if err != nil {
		return err
	}

	password, err := domain.NewPasswordFromRaw(cmd.Password)
	if err != nil {
		return err
	}

	var opts []domain.UserOption

	if cmd.Email != nil {
		email, err := domain.NewEmail(*cmd.Email)
		if err != nil {
			return err
		}
		opts = append(opts, domain.WithEmail(email))
	}

	if cmd.Username != nil {
		opts = append(opts, domain.WithUsername(*cmd.Username))
	}

	// NewUser sets isApproved=false by default.
	user := domain.NewUser(phone, password, opts...)

	if err := h.repo.Save(ctx, user); err != nil {
		h.logger.Errorf("failed to save user during sign-up: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish sign-up events: %v", err)
	}

	return nil
}
