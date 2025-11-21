package sprites

import "github.com/gin-gonic/gin"

// ModelsController defines the interface for model-related HTTP operations.
type ModelsController interface {

	// DownloadSpriteData handles the download of sprite file.
	DownloadModelsByWorldId(c *gin.Context)

	// UploadSpriteData handles the upload of sprite file.
	UploadModelsByWorldId(c *gin.Context)
}
