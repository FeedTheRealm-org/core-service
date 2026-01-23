package middleware

import (
	"strings"
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware parses the JWT token included in the request header,
// and populates the gin context with it. If it cant find the header or
// cant decode the token it passes to the next middleware without setting anything.
func JWTAuthMiddleware(jwtManager *session.JWTManager, fixedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			return
		}
		c.Set("includedJWT", true)

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// If a fixed server token is configured and matches the provided token,
		// treat it as a valid session without further validation.
		// set a dummy userID so that helpers like
		// GetUserIDFromSession don't fail when parsing it, which allows
		// read-only endpoints that only check for a valid session to work.
		if fixedToken != "" && tokenString == fixedToken {
			c.Set("userID", "00000000-0000-0000-0000-000000000000")
			return
		}

		claims, err := jwtManager.IsValidateToken(tokenString, time.Now())
		if err != nil {
			if _, ok := err.(*session.JWTExpiredTokenError); ok {
				c.Set("expiredJWT", true)
			}
			logger.Logger.Warnf("Invalid JWT")
			c.Set("invalidJWT", true)
			return
		}

		if userID, ok := claims["userID"].(string); ok {
			c.Set("userID", userID)
		} else {
			logger.Logger.Warnln("Missing userID in JWT claims")
			c.Set("invalidJWT", true)
		}
	}
}
