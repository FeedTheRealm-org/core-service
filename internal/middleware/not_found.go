package middleware

import (
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
)

func NotFoundController(c *gin.Context) {
	_ = c.Error(errors.NewNotFoundError("The requested resource was not found"))
}
