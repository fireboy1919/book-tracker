//go:build serverless
// +build serverless

package main

// This file is used when building for serverless mode (Vercel)
// The actual handler is in api/handler.go
func main() {
	// This main function should not be called in serverless mode
	// The Handler function in api/handler.go is the entry point
}