package usecase

import (
	"context"
)

type UserUseCaseI interface {
	// Add methods as needed or just UseCaseI if defined in sub-package
}

type AuthzUseCaseI interface {
	// ...
}

type AuditUseCaseI interface {
	// ...
}

type MinioUseCaseI interface {
	UploadImage(ctx context.Context, imageFile any, imageSize int64, contextType string) (string, error)
	// ... add other methods
}

// Actually, I should probably check if sub-packages already have these interfaces and use them.
