package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware parses the JWT token included in the request header,
// and populates the gin context with it. If it cant find the header or
// cant decode the token it passes to the next middleware without setting anything.
func JWTAuthMiddleware(jwtManager *session.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		c.Set("includedJWT", true)

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtManager.IsValidateToken(tokenString, time.Now())
		if err != nil {
			if errors.Is(err, &session.JWTExpiredTokenError{}) {
				logger.Logger.Infoln("JWT token has expired")
				c.Set("expiredJWT", true)
			}
			logger.Logger.Infoln("JWT token is invalid")
			c.Set("invalidJWT", true)
		}

		if userID, ok := claims["userID"].(string); ok {
			c.Set("userID", userID)
		} else {
			logger.Logger.Warnln("Missing userID in JWT claims")
			c.Set("invalidJWT", true)
		}

		c.Next()
	}
}
