//go:build production
// +build production

package main

import "github.com/gin-gonic/gin"

func setupTestRoutes(api *gin.RouterGroup) {
	// No test routes in production builds
}