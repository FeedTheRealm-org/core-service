package middleware

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/internal/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware is a Gin middleware that handles errors
// that occur during the request lifecycle, in a centralized manner.
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Execute after everything else

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err, ok := c.Errors.Last().Err.(*errors.HttpError)
		if !ok {
			err = errors.NewInternalServerError("An unexpected error occurred")
		}

		errorResponse := dtos.ErrorResponse{
			Type:     "about:blank",
			Title:    http.StatusText(err.Status),
			Status:   err.Status,
			Detail:   err.Message,
			Instance: c.Request.RequestURI,
		}

		c.JSON(err.Status, errorResponse)
		c.Writer.WriteHeaderNow()
	}
}
