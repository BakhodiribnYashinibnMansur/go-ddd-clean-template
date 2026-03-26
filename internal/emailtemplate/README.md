# Email Template

Manages reusable email templates with support for HTML and plain-text bodies, dynamic template variables, and soft-delete capability. Templates are used by other bounded contexts (e.g., notifications) to render and send emails.

## Domain

### Aggregate Root
- `EmailTemplate` -- Represents a single email template. Key fields: `name`, `subject`, `htmlBody`, `textBody`, `variables` (list of placeholder names). Embeds `shared.AggregateRoot` with soft-delete support (`deletedAt`).

### Domain Events
- `TemplateUpdated` -- Raised when an email template is updated. Carries the aggregate ID and template `Name`.

### Domain Errors
- `ErrEmailTemplateNotFound` -- Returned when an email template cannot be found by ID.

### Repository Interfaces
- `EmailTemplateRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`
- `EmailTemplateReadRepository` (read) -- `FindByID`, `List` (returns `EmailTemplateView` projections)

### Filter
- `EmailTemplateFilter` -- Supports text `Search` plus `Limit`/`Offset` pagination.

## Application (CQRS)

### Commands
- `CreateCommand` / `CreateHandler` -- Creates a new email template with name, subject, HTML body, optional text body, and variable placeholders.
- `UpdateCommand` / `UpdateHandler` -- Partially updates an existing template (all fields optional via pointers, except `Variables` which replaces the full list when provided). Raises `TemplateUpdated` event.
- `DeleteCommand` / `DeleteHandler` -- Deletes an email template by ID.

### Queries
- `GetQuery` / `GetHandler` -- Fetches a single email template view by ID.
- `ListQuery` / `ListHandler` -- Returns a paginated, searchable list of email template views with total count.

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/email-templates` | Create a new email template |
| GET | `/email-templates` | List email templates (paginated: `limit`, `offset` query params) |
| GET | `/email-templates/:id` | Get a single email template by ID |
| PATCH | `/email-templates/:id` | Partially update an email template |
| DELETE | `/email-templates/:id` | Delete an email template |

## Usage
```go
import "gct/internal/emailtemplate"
```
