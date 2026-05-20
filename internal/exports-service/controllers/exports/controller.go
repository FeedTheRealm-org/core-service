package exports

import "github.com/gin-gonic/gin"

// ExportsController defines the interface for export-related HTTP operations.
type ExportsController interface {
	UploadZip(c *gin.Context)
	GetZipPath(c *gin.Context)
	ListZipVersions(c *gin.Context)
	DeleteZipVersion(c *gin.Context)
	SetLatestZipVersion(c *gin.Context)
}
