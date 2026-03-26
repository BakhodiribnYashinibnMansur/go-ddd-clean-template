package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EmailTemplateFilter carries optional criteria for querying email templates.
// Search performs a substring match against the template name.
type EmailTemplateFilter struct {
	Search *string
	Limit  int64
	Offset int64
}

// EmailTemplateView is a read-model DTO for email templates.
type EmailTemplateView struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	HTMLBody  string    `json:"html_body"`
	TextBody  string    `json:"text_body"`
	Variables []string  `json:"variables"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmailTemplateRepository is the write-side repository for the EmailTemplate aggregate.
// Implementations must return ErrEmailTemplateNotFound from FindByID when no row matches.
type EmailTemplateRepository interface {
	Save(ctx context.Context, entity *EmailTemplate) error
	FindByID(ctx context.Context, id uuid.UUID) (*EmailTemplate, error)
	Update(ctx context.Context, entity *EmailTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// EmailTemplateReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected EmailTemplateView DTOs for list and detail queries.
type EmailTemplateReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*EmailTemplateView, error)
	List(ctx context.Context, filter EmailTemplateFilter) ([]*EmailTemplateView, int64, error)
}
