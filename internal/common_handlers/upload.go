package common_handlers

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

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
