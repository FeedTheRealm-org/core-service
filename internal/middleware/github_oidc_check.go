package middleware

import (
	"io"
	"strings"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/oidc_validation"
	"github.com/gin-gonic/gin"
)

// GithubOIDCCheck is a Gin middleware that checks if the request has a valid GitHub OIDC token
// and sets the "invalidGithubOIDC" flag in the context accordingly.
func GithubOIDCCheck(ghv *oidc_validation.GitHubOIDCVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()

		// To be deleted
		rawBody, _ := io.ReadAll(c.Request.Body)
		logger.Logger.Infow("request log",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"request_body", string(rawBody),
		)

		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ghv.IsValidToken(tokenString)
		if err != nil {
			c.Set("invalidGithubOIDC", true)
			return
		}

		if !ghv.IsValidRepo(&claims) {
			c.Set("invalidGithubOIDC", true)
			return
		}

		if !ghv.IsTriggerATag(&claims) {
			c.Set("invalidGithubOIDC", true)
			return
		}

		c.Set("invalidGithubOIDC", false)
	}
}
