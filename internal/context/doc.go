// Package contexts groups all bounded contexts (Evans DDD) under core/,
// supporting/, and generic/ subdomain tiers. No bounded context may import
// another bounded context directly; communication flows only through
// gct/internal/contract/events (Published Language) and
// gct/internal/contract/ports (Anti-Corruption Layer interfaces).
package contexts
