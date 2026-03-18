package middleware

import (
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// AdminCheckMiddleware will check if the user is admin before allowing swagger endpoints
func AdminCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := common_handlers.IsAdminSession(c); err != nil {
			c.Abort()
			_ = c.Error(errors.NewForbiddenError(err.Error()))
			return
		}

		c.Next()
	}
}
