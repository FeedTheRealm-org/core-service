package middleware

import (
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/gin-gonic/gin"
)

// MultipartCleanupMiddleware removes multipart temp files after the request completes.
func MultipartCleanupMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		common_handlers.CleanupMultipartRequest(c)
	}
}
