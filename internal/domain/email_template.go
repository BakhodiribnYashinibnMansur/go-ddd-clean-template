package domain

import "time"

type EmailTemplate struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Subject   string    `json:"subject" db:"subject"`
	HtmlBody  string    `json:"html_body" db:"html_body"`
	TextBody  string    `json:"text_body" db:"text_body"`
	Type      string    `json:"type" db:"type"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type EmailTemplateFilter struct {
	Search string
	Type   string
	Limit  int
	Offset int
}

type CreateEmailTemplateRequest struct {
	Name     string `json:"name" binding:"required"`
	Subject  string `json:"subject" binding:"required"`
	HtmlBody string `json:"html_body" binding:"required"`
	TextBody string `json:"text_body"`
	Type     string `json:"type" binding:"required"`
}

type UpdateEmailTemplateRequest struct {
	Name     *string `json:"name"`
	Subject  *string `json:"subject"`
	HtmlBody *string `json:"html_body"`
	TextBody *string `json:"text_body"`
	Type     *string `json:"type"`
	IsActive *bool   `json:"is_active"`
}

