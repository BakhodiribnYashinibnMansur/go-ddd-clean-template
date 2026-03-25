package application

import (
	"time"

	"github.com/google/uuid"
)

// EmailTemplateView is a read-model DTO returned by query handlers.
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
