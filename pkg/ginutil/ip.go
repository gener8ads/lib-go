package ginutil

import (
	"github.com/gin-gonic/gin"
)

// ExtractIP address from a request context
func ExtractIP(c *gin.Context) string {
	cfIP := c.Request.Header.Get("cf-connecting-ip")
	forwarded := c.Request.Header.Get("x-forwarded-for")

	switch {
	case cfIP != "":
		return cfIP

	case forwarded != "":
		return forwarded

	default:
		return c.Request.RemoteAddr
	}
}
