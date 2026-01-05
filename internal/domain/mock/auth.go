package mock

import (
	"gct/internal/domain"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

// SignInIn generates a fake domain.SignInIn
func SignInIn() *domain.SignInIn {
	return &domain.SignInIn{
		Phone:     Phone(),
		Password:  gofakeit.Password(true, true, true, true, false, 12),
		DeviceID:  UUID(),
		IP:        gofakeit.IPv4Address(),
		UserAgent: gofakeit.UserAgent(),
	}
}

// SignInInWithPhone generates a fake domain.SignInIn with specific phone
func SignInInWithPhone(phone string) *domain.SignInIn {
	signIn := SignInIn()
	signIn.Phone = phone
	return signIn
}

// SignInOut generates a fake domain.SignInOut
func SignInOut() *domain.SignInOut {
	return &domain.SignInOut{
		AccessToken:  gofakeit.LetterN(64),
		RefreshToken: gofakeit.LetterN(64),
	}
}

// SignUpIn generates a fake domain.SignUpIn
func SignUpIn() *domain.SignUpIn {
	return &domain.SignUpIn{
		Phone:     Phone(),
		Password:  gofakeit.Password(true, true, true, true, false, 12),
		Username:  gofakeit.Name(),
		Email:     gofakeit.Email(),
		DeviceID:  UUID(),
		IP:        gofakeit.IPv4Address(),
		UserAgent: gofakeit.UserAgent(),
	}
}

// SignUpInWithEmail generates a fake domain.SignUpIn with specific email
func SignUpInWithEmail(email string) *domain.SignUpIn {
	signUp := SignUpIn()
	signUp.Email = email
	return signUp
}

// SignOutIn generates a fake domain.SignOutIn
func SignOutIn() *domain.SignOutIn {
	return &domain.SignOutIn{
		SessionID: UUID(),
		UserID:    UUID(),
	}
}

// SignOutInWithUserID generates a fake domain.SignOutIn with specific user ID
func SignOutInWithUserID(userID uuid.UUID) *domain.SignOutIn {
	signOut := SignOutIn()
	signOut.UserID = userID
	return signOut
}

// RefreshIn generates a fake domain.RefreshIn
func RefreshIn() *domain.RefreshIn {
	return &domain.RefreshIn{
		SessionID: UUID(),
	}
}
