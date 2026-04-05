package consts

// Cache configuration. CacheInvalidationChannel is the Redis pub/sub channel name used
// to broadcast cache busts across multiple application instances in a horizontally scaled deployment.
const (
	DefaultCacheCapacity     = 100
	CacheInvalidationChannel = "cache_invalidation"

	// Cache keys
	CacheKeyPrefix     = "cache_"
	CacheUserKey       = TableUsers
	CacheRoleKey       = TableRole
	CachePermissionKey = TablePermission
	CachePolicyKey     = TablePolicy
	CacheSessionKey    = TableSession
	CacheRelationKey   = TableRelation
)
