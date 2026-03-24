package consts

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
