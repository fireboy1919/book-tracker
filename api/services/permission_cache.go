package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PermissionCacheEntry represents a cached permission result
type PermissionCacheEntry struct {
	HasPermission bool
	ExpiresAt     time.Time
}

// PermissionCache provides request-scoped permission caching
type PermissionCache struct {
	cache map[string]PermissionCacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewPermissionCache creates a new permission cache with TTL in milliseconds
func NewPermissionCache(ttlMs int64) *PermissionCache {
	return &PermissionCache{
		cache: make(map[string]PermissionCacheEntry),
		ttl:   time.Duration(ttlMs) * time.Millisecond,
	}
}

// GetOrCheck gets permission from cache or checks and caches the result
func (pc *PermissionCache) GetOrCheck(userID uint, childID uint, permissionType string) (bool, error) {
	key := fmt.Sprintf("%d:%d:%s", userID, childID, permissionType)
	
	// Try to get from cache first
	pc.mutex.RLock()
	if entry, exists := pc.cache[key]; exists && time.Now().Before(entry.ExpiresAt) {
		pc.mutex.RUnlock()
		return entry.HasPermission, nil
	}
	pc.mutex.RUnlock()
	
	// Not in cache or expired, check permission
	hasPermission, err := CheckChildPermission(userID, childID, permissionType)
	if err != nil {
		return false, err
	}
	
	// Cache the result
	pc.mutex.Lock()
	pc.cache[key] = PermissionCacheEntry{
		HasPermission: hasPermission,
		ExpiresAt:     time.Now().Add(pc.ttl),
	}
	pc.mutex.Unlock()
	
	return hasPermission, nil
}

// Clear removes all cached entries
func (pc *PermissionCache) Clear() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.cache = make(map[string]PermissionCacheEntry)
}

// Context key for permission cache
type contextKey string

const permissionCacheKey contextKey = "permissionCache"

// GetPermissionCacheFromContext gets or creates permission cache from context
func GetPermissionCacheFromContext(ctx context.Context) *PermissionCache {
	if cache, ok := ctx.Value(permissionCacheKey).(*PermissionCache); ok {
		return cache
	}
	// Create new cache with 5 minute TTL for request duration
	return NewPermissionCache(5 * 60 * 1000)
}

// SetPermissionCacheInContext sets permission cache in context
func SetPermissionCacheInContext(ctx context.Context, cache *PermissionCache) context.Context {
	return context.WithValue(ctx, permissionCacheKey, cache)
}