package consts

const (
	CtxUserID           string = "user_id"
	CtxCompanyID        string = "company_id"
	CtxRoleID           string = "role_id"
	CtxRoleTitle        string = "role_title"
	CtxDeviceID         string = "device_id"
	CtxSessionID        string = "session_id"
	CtxSession          string = "session_object"
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
)
