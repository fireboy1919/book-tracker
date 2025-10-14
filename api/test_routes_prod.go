//go:build production
// +build production

package handler

import "github.com/gin-gonic/gin"

func setupTestRoutes(api *gin.RouterGroup) {
	// No test routes in production builds
}