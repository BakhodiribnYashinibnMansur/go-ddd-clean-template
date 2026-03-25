package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// EmailTemplate is the aggregate root for email template management.
type EmailTemplate struct {
	shared.AggregateRoot
	name      string
	subject   string
	htmlBody  string
	textBody  string
	variables []string
}

// NewEmailTemplate creates a new EmailTemplate aggregate.
func NewEmailTemplate(name, subject, htmlBody, textBody string, variables []string) *EmailTemplate {
	if variables == nil {
		variables = make([]string, 0)
	}
	et := &EmailTemplate{
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		subject:       subject,
		htmlBody:      htmlBody,
		textBody:      textBody,
		variables:     variables,
	}
	return et
}

// ReconstructEmailTemplate rebuilds an EmailTemplate aggregate from persisted data. No events are raised.
func ReconstructEmailTemplate(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, subject, htmlBody, textBody string,
	variables []string,
) *EmailTemplate {
	if variables == nil {
		variables = make([]string, 0)
	}
	return &EmailTemplate{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:          name,
		subject:       subject,
		htmlBody:      htmlBody,
		textBody:      textBody,
		variables:     variables,
	}
}

// UpdateDetails updates mutable fields and raises a TemplateUpdated event.
func (et *EmailTemplate) UpdateDetails(name, subject, htmlBody, textBody *string, variables []string) {
	if name != nil {
		et.name = *name
	}
	if subject != nil {
		et.subject = *subject
	}
	if htmlBody != nil {
		et.htmlBody = *htmlBody
	}
	if textBody != nil {
		et.textBody = *textBody
	}
	if variables != nil {
		et.variables = variables
	}
	et.Touch()
	et.AddEvent(NewTemplateUpdated(et.ID(), et.name))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (et *EmailTemplate) Name() string        { return et.name }
func (et *EmailTemplate) Subject() string     { return et.subject }
func (et *EmailTemplate) HTMLBody() string    { return et.htmlBody }
func (et *EmailTemplate) TextBody() string    { return et.textBody }
func (et *EmailTemplate) Variables() []string { return et.variables }
