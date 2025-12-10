package bundles

import "github.com/gin-gonic/gin"

// BundleController defines the interface for model-related HTTP operations.
type BundleController interface {

	// DownloadSpriteData handles the download of sprite file.
	DownloadWorldBundle(c *gin.Context)

	// UploadSpriteData handles the upload of sprite file.
	UploadWorldBundle(c *gin.Context)
}
