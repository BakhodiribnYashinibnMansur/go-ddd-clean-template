package http

// CreateRequest represents the request body for creating an email template.
type CreateRequest struct {
	Name      string   `json:"name" binding:"required"`
	Subject   string   `json:"subject" binding:"required"`
	HTMLBody  string   `json:"html_body" binding:"required"`
	TextBody  string   `json:"text_body"`
	Variables []string `json:"variables,omitempty"`
}

// UpdateRequest represents the request body for updating an email template.
type UpdateRequest struct {
	Name      *string  `json:"name,omitempty"`
	Subject   *string  `json:"subject,omitempty"`
	HTMLBody  *string  `json:"html_body,omitempty"`
	TextBody  *string  `json:"text_body,omitempty"`
	Variables []string `json:"variables,omitempty"`
}
