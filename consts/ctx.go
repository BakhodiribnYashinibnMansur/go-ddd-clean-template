package consts

const (
	CtxUserId                  = "user_id"
	CtxUserID                  = "id"
	CtxAbonentId               = "abonent_id"
	CtxCompanyID               = "company_id"
	CtxRoleID                  = "role_id"
	CtxStaffID                 = "staff_id"
	CtxSessionID               = "session_id"
	CtxRoleTitle               = "role_title"
	CtxExpiredDate             = "expired_date"
	CtxIsAbonent               = "is_abonent"
	CtxIsCorp                  = "is_corp"
	AuthorizationHeader        = "Authorization"
	BasicToken                 = "Basic"
	BearerToken                = "Bearer"
	TokenAccessType     string = "access"
	TokenRefreshType    string = "refresh"

	// Claims
	ClaimIssuer          = "iss" // Issuer
	ClaimSubject         = "sub" // Subject
	ClaimSessionID       = "sid" // Session ID
	ClaimCompanyID       = "cid" // Company ID
	ClaimAudience        = "aud" // Audience
	ClaimScope           = "scp" // Scope
	ClaimAuthorizedParty = "azp" // Authorized Party
	ClaimType            = "typ" // Type
	ClaimExpiresAt       = "exp" // Expires At
	ClaimIssuedAt        = "iat" // Issued At
	ClaimJWTID           = "jti" // JWT ID
)
