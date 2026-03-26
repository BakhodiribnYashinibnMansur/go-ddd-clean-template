package consts

// Context keys injected by authentication/authorization middleware and read by handlers.
// These are string-typed to work with gin.Context.Get/Set. Never use raw string literals
// elsewhere — always reference these constants to enable compile-time refactoring.
const (
	CtxUserID           string = "user_id"
	CtxUser             string = "user"
	CtxCompanyID        string = "company_id"
	CtxRoleID           string = "role_id"
	CtxRoleTitle        string = "role_title"
	CtxDeviceID         string = "device_id"
	CtxSessionID        string = "session_id"
	CtxRefreshToken     string = "refresh_token"
	CtxSession          string = "session_data"
	CtxMockMode         string = "mock_mode"
	CtxIsAdmin          string = "is_admin"
	CtxApiKeyAuth       string = "api_key_authenticated"
	CtxExpiredDate      string = "expired_date"
	CtxIsAbonent        string = "is_abonent"
	CtxIsCorp           string = "is_corp"
	CtxStaffID          string = "staff_id"
	AuthorizationHeader string = "Authorization"
	BasicToken          string = "Basic"
	BearerToken         string = "Bearer"
	TokenAccessType     string = "access"
	TokenRefreshType    string = "refresh"

	// Claims
	ClaimIssuer          string = "iss" // Issuer
	ClaimSubject         string = "sub" // Subject
	ClaimSessionID       string = "sid" // Session ID
	ClaimCompanyID       string = "cid" // Company ID
	ClaimAudience        string = "aud" // Audience
	ClaimScope           string = "scp" // Scope
	ClaimAuthorizedParty string = "azp" // Authorized Party
	ClaimType            string = "typ" // Type
	ClaimExpiresAt       string = "exp" // Expires At
	ClaimIssuedAt        string = "iat" // Issued At
	ClaimJWTID           string = "jti" // JWT ID

	// Integration Context Keys
	CtxIntegrationID   string = "integration_id"
	CtxAPIKeyID        string = "api_key_id"
	CtxIntegrationName string = "integration_name"
)

// Middleware infrastructure constants: request tracing, status thresholds, and timeout values.
const (
	// Context keys
	CtxKeyRequestID string = "request_id"

	// HTTP status code thresholds
	HTTPStatusSuccessThreshold = 400 // Status codes below this are considered successful

	// Timeouts
	AuditPersistTimeout = 5 // seconds for audit log persistence
)
