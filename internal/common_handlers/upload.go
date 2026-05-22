package common_handlers

import (
	"mime/multipart"

	internalErrors "github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/gin-gonic/gin"
)

// PrepareMultipartRequest parses a multipart request using a bounded in-memory buffer.
func PrepareMultipartRequest(c *gin.Context) error {
	if _, err := c.MultipartForm(); err != nil {
		return internalErrors.NewBadRequestError("failed to parse multipart form")
	}

	return nil
}

// CleanupMultipartRequest removes any temporary multipart files created while parsing the request.
func CleanupMultipartRequest(c *gin.Context) {
	if c.Request != nil && c.Request.MultipartForm != nil {
		_ = c.Request.MultipartForm.RemoveAll()
	}
}

// OpenMultipartFile safely opens a multipart file header.
func OpenMultipartFile(fileHeader *multipart.FileHeader) (multipart.File, error) {
	return fileHeader.Open()
}
