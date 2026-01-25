package contextx

import "context"

type ctxKey string

const (
	FieldRequestID  = "request_id"
	FieldSessionID  = "session_id"
	FieldUserID     = "user_id"
	FieldUserRole   = "user_role"
	FieldIPAddress  = "ip_address"
	FieldUserAgent  = "user_agent"
	FieldAPIVersion = "api_version"
	FieldTraceID    = "trace_id"
)

const (
	requestIDKey  ctxKey = FieldRequestID
	sessionIDKey  ctxKey = FieldSessionID
	userIDKey     ctxKey = FieldUserID
	userRoleKey   ctxKey = FieldUserRole
	ipAddressKey  ctxKey = FieldIPAddress
	userAgentKey  ctxKey = FieldUserAgent
	apiVersionKey ctxKey = FieldAPIVersion
	traceIDKey    ctxKey = FieldTraceID
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func GetSessionID(ctx context.Context) string {
	id, _ := ctx.Value(sessionIDKey).(string)
	return id
}

func WithUserID(ctx context.Context, id interface{}) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func GetUserID(ctx context.Context) interface{} {
	return ctx.Value(userIDKey)
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(userRoleKey).(string)
	return role
}

func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipAddressKey, ip)
}

func GetIPAddress(ctx context.Context) string {
	ip, _ := ctx.Value(ipAddressKey).(string)
	return ip
}

func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, userAgentKey, ua)
}

func GetUserAgent(ctx context.Context) string {
	ua, _ := ctx.Value(userAgentKey).(string)
	return ua
}

func WithAPIVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersionKey, version)
}

func GetAPIVersion(ctx context.Context) string {
	v, _ := ctx.Value(apiVersionKey).(string)
	return v
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

func GetTraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey).(string)
	return id
}
