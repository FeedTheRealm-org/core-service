package sprites

import "github.com/gin-gonic/gin"

// ModelsController defines the interface for model-related HTTP operations.
type ModelsController interface {

	// DownloadSpriteData handles the download of sprite file.
	DownloadModel(c *gin.Context)

	// RetrieveWorldAssetIds retrieves asset IDs associated with a specific world.
	ListAssets(c *gin.Context)

	// UploadSpriteData handles the upload of sprite file.
	UploadModels(c *gin.Context)
}
