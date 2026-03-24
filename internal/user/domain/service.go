package domain

// SignInService is a stateless domain service that orchestrates sign-in logic.
type SignInService struct{}

// SignIn validates credentials and creates a session on the User aggregate.
// It does NOT touch any infrastructure — pure domain logic.
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
