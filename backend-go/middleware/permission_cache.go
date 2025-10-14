package middleware

import (
	"github.com/booktracker/backend-go/services"
	"github.com/gin-gonic/gin"
)

// PermissionCacheMiddleware adds permission cache to request context
func PermissionCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create new permission cache for this request
		cache := services.NewPermissionCache(5 * 60 * 1000) // 5 minutes in milliseconds
		
		// Store in gin context for easy access
		c.Set("permissionCache", cache)
		
		c.Next()
	}
}

// GetPermissionCache gets permission cache from gin context
func GetPermissionCache(c *gin.Context) *services.PermissionCache {
	if cache, exists := c.Get("permissionCache"); exists {
		if permCache, ok := cache.(*services.PermissionCache); ok {
			return permCache
		}
	}
	// Fallback to new cache if not found
	return services.NewPermissionCache(5 * 60 * 1000)
}