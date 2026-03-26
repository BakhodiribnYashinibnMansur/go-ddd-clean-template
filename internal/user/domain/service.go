package domain

// SignInService is a stateless domain service that coordinates the multi-step sign-in flow.
// It exists because sign-in spans multiple aggregate concerns (status checks, password verification,
// session creation) that don't naturally belong to a single method on User.
type SignInService struct{}

// SignIn validates credentials and creates a session on the User aggregate.
// Preconditions: user must be active AND approved. Returns domain errors (ErrUserInactive,
// ErrUserNotApproved, ErrInvalidPassword, ErrMaxSessionsReached) — no infrastructure side effects.
func (s *SignInService) SignIn(
	user *User,
	rawPassword string,
	deviceType SessionDeviceType,
	ip, userAgent string,
) (*Session, error) {
	if !user.IsActive() {
		return nil, ErrUserInactive
	}
	if !user.IsApproved() {
		return nil, ErrUserNotApproved
	}
	if err := user.VerifyPassword(rawPassword); err != nil {
		return nil, err
	}
	session, err := user.AddSession(deviceType, ip, userAgent)
	if err != nil {
		return nil, err
	}
	user.UpdateLastSeen()
	return session, nil
}
