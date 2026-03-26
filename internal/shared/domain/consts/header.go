// Package consts defines application-wide constants shared across all bounded contexts.
// These values are intentionally kept in a single place to prevent drift between layers.
package consts

// HTTP header names used by middleware, handlers, and CORS configuration.
// Security headers (HSTS, CSP, X-Frame-Options) are set by the security middleware on every response.
const (
	HeaderAuthorization                 = "Authorization"
	HeaderXRequestID                    = "X-Request-ID"
	HeaderXDeviceID                     = "X-Device-ID"
	HeaderXAPIKey                       = "X-API-KEY"
	HeaderXTimeUnix                     = "X-Time-Unix"
	HeaderXSign                         = "X-Sign"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderOrigin                        = "Origin"
	HeaderUserAgent                     = "User-Agent"
	HeaderAcceptLanguage                = "Accept-Language"
	HeaderLanguage                      = "Language"
	HeaderXApiKeyType                   = "X-Api-Key-Type"
	HeaderAppVersion                    = "appVersion"
	HeaderContentType                   = "Content-Type"
	HeaderXTotalCount                   = "X-Total-Count"
	HeaderXCSRFToken                    = "X-CSRF-Token"
	HeaderCacheControl                  = "Cache-Control"
	HeaderXRequestedWith                = "X-Requested-With"
	HeaderXContentTypeOptions           = "X-Content-Type-Options"
	HeaderXFrameOptions                 = "X-Frame-Options"
	HeaderXXSSProtection                = "X-XSS-Protection"
	HeaderReferrerPolicy                = "Referrer-Policy"
	HeaderXPermittedCP                  = "X-Permitted-Cross-Domain-Policies"
	HeaderCSP                           = "Content-Security-Policy"
	HeaderHSTS                          = "Strict-Transport-Security"
	HeaderSecFetchSite                  = "Sec-Fetch-Site"
	HeaderSecFetchMode                  = "Sec-Fetch-Mode"
	HeaderSecFetchDest                  = "Sec-Fetch-Dest"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	ParamAPIKey                         = "api_key"
)

// Standard header values for security-related response headers.
const (
	HeaderValueNoSniff      = "nosniff"
	HeaderValueDeny         = "DENY"
	HeaderValueXSSBlock     = "1; mode=block"
	HeaderValueStrictOrigin = "strict-origin-when-cross-origin"
	HeaderValueNone         = "none"
	HeaderValueHSTS         = "max-age=31536000; includeSubDomains; preload"
	HeaderValueSameOrigin   = "same-origin"
	HeaderValueSameSite     = "same-site"
	HeaderValueNavigate     = "navigate"
	HeaderValueDocument     = "document"
)
