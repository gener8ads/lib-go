package ginutil

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-version"
)

// ExtractClient request metadata
func ExtractClient(c *gin.Context) (name, version string) {
	raw := c.Request.Header.Get("x-client")
	parts := strings.Split(raw, "/")

	if len(parts) > 0 {
		name = parts[0]
	}

	if len(parts) >= 2 {
		version = parts[1]
	}

	return name, version
}

// MinVersion middleware
func MinVersion(minVerStr string, code int) gin.HandlerFunc {
	minVersion, err := version.NewVersion(minVerStr)

	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		_, verStr := ExtractClient(c)

		if verStr == "" {
			c.AbortWithStatus(code)
			return
		}

		clientVer, verErr := version.NewVersion(verStr)

		if verErr != nil || clientVer.LessThan(minVersion) {
			c.AbortWithStatus(code)
		} else {
			c.Next()
		}
	}
}
