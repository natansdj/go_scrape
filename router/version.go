package router

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
)

var version string

// SetVersion for setup version string.
func SetVersion(ver string) {
	version = ver
}

// GetVersion for get current version.
func GetVersion() string {
	return version
}

// PrintVersion provide print server engine
func PrintVersion() {
	fmt.Printf(`GoScrape %s, Compiler: %s %s, Copyright (C) 2021 Nath.`,
		version,
		runtime.Compiler,
		runtime.Version())
	fmt.Println()
}

// VersionMiddleware : add version on header.
func VersionMiddleware() gin.HandlerFunc {
	// Set out header value for each response
	return func(c *gin.Context) {
		c.Header("X-GOSCRAPE-VERSION", version)
		c.Next()
	}
}
