package middleware

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowed := range conf.CORSAllowedOrigins {
			if allowed == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				break
			}

			if origin == allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
