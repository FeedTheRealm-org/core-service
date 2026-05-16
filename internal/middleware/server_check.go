package middleware

import (
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// ServerCheckMiddleware will check if the user is server before allowing swagger endpoints
func ServerCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := common_handlers.IsServerSession(c); err != nil {
			c.Abort()
			_ = c.Error(errors.NewForbiddenError(err.Error()))
			return
		}

		c.Next()
	}
}
